execute('Wait for K8s cluster to come up') do
  Chef::Log.info("Wait for K8s cluster to come up, there should be #{node['k8sNodeCount']} number of nodes")
  action 'run'
  retries 30
  retry_delay 10
  command "kubectl get nodes --kubeconfig=/home/ubuntu/.kube/config| grep ' Ready' | wc -l | grep -w #{node['k8sNodeCount']}"
  returns 0
end

execute('Remove master taint') do
  action 'run'
  retries 2
  retry_delay 2
  command 'kubectl taint nodes -l node-role.kubernetes.io/master node-role.kubernetes.io/master:NoSchedule- --kubeconfig=/home/ubuntu/.kube/config'
  returns 0
  ignore_failure true
end

execute('Assign k8s node labels for master') do
  action 'run'
  retries 2
  retry_delay 2
  Chef::Log.info("Setting label platform-cluster-master for #{node['platform-cluster-master']} ")
  command "kubectl label nodes #{node['platform-cluster-master']} harole=master --kubeconfig=/home/ubuntu/.kube/config"
  returns 0
  ignore_failure true
end

execute('Assign k8s node labels for primary') do
  action 'run'
  retries 2
  retry_delay 2
  Chef::Log.info("Setting label platform-cluster-primary-node for #{node['platform-cluster-primary-node']} ")
  command "kubectl label nodes #{node['platform-cluster-primary-node']} harole=primary --kubeconfig=/home/ubuntu/.kube/config"
  returns 0
  ignore_failure true
end

execute('Assign k8s node labels for secondary') do
  action 'run'
  retries 2
  retry_delay 2
  Chef::Log.info("Setting label platform-cluster-secondary-node for #{node['platform-cluster-secondary-node']} ")
  command "kubectl label nodes #{node['platform-cluster-secondary-node']} harole=secondary --kubeconfig=/home/ubuntu/.kube/config"
  returns 0
  ignore_failure true
end

execute('Setup docker registry secrets') do
  action 'run'
  retries 2
  retry_delay 2
  regsecrets = data_bag_item('mexsecrets', 'docker_registry')
  Chef::Log.info("Create secret to access #{node['edgeCloudImage']} as user #{regsecrets['mex_docker_username']}")
  command "kubectl create secret docker-registry mexreg-secret --docker-server=#{node['edgeCloudImage']} --docker-username=#{regsecrets['mex_docker_username']} --docker-password=#{regsecrets['mex_docker_password']} --docker-email=mobiledgex.ops@mobiledgex.com --kubeconfig=/home/ubuntu/.kube/config"
  returns 0
  ignore_failure true
end

# start redis on master if HA enabled
template '/home/ubuntu/k8s-deployment-redis.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'master',
     deploymentName: 'redis',
     version: node['redisVersion'],
     headlessSvcs: {
          redis: {
              serviceName: node['redisServiceName'],
              ports: { redisServicePort: { protocol: 'TCP', portNum: node['redisServicePort'] } },
              appSelector: 'redis',
          },
     },
     services: {
          redis: {
              image: node['redisImage'] + ':' + node['redisVersion'],
              serviceName: node['redisServiceName'],
              port: node['redisServicePort'],
              env: [ 'ALLOW_EMPTY_PASSWORD=yes' ],
          },
     },
     hostvols: {},
     configmaps: {}
   )
  only_if { node.attribute?(:redisServicePort) }
end

execute('Setup redis deployment') do
  action 'run'
  command 'kubectl apply -f /home/ubuntu/k8s-deployment-redis.yaml --kubeconfig=/home/ubuntu/.kube/config'
  returns 0
  only_if { node.attribute?(:redisServicePort) }
end

execute('Wait for redis deployment to come up') do
  Chef::Log.info('Wait for redis deployment to come up')
  action 'run'
  retries 20
  retry_delay 6
  command 'kubectl get pods -l app=redis -l version=' + node['redisVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
  returns 0
  only_if { node.attribute?(:redisServicePort) }
end

svc_vars_primary = get_services_vars('primary')
svc_vars_secondary = get_services_vars('secondary')
hostvol_vars = get_hostvols_vars
configmap_vars = get_configmap_vars

cookbook_file '/home/ubuntu/prometheus.yml' do
  source 'prometheus.yml'
  mode '0644'
  action :create_if_missing
  force_unlink true
  notifies :run, 'execute[create-prometheus-configmap]', :immediately
end

execute('create-prometheus-configmap') do
  Chef::Log.info('create prometheus configmap')
  action :nothing
  command 'kubectl create configmap prom-cm --from-file prometheus.yml=/home/ubuntu/prometheus.yml --kubeconfig=/home/ubuntu/.kube/config'
  retries 2
  retry_delay 2
  returns 0
  ignore_failure true
end

# start processes only any node if no HA enabled
template '/home/ubuntu/k8s-deployment.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'simplex',
     deploymentName: 'platform-simplex',
     version: node['edgeCloudVersion'],
     services: svc_vars_primary,
     hostvols: hostvol_vars,
     configmaps: configmap_vars
   )
  not_if { node.attribute?(:redisServicePort) }
end

execute('Setup simplex deployment') do
  action 'run'
  command 'kubectl apply -f /home/ubuntu/k8s-deployment.yaml --kubeconfig=/home/ubuntu/.kube/config'
  returns 0
  not_if { node.attribute?(:redisServicePort) }
end

execute('Wait for simplex platform pod to come up') do
  Chef::Log.info('Wait for simplex platform pod to come up')
  action 'run'
  retries 30
  retry_delay 10 
  command 'kubectl get pods -l app=platform-simplex -l version=' + node['edgeCloudVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
  returns 0
  not_if { node.attribute?(:redisServicePort) }
end

# update primary and secondary manifests. Redeploy primary and then secondary if there are changes

template '/home/ubuntu/k8s-deployment-primary.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'primary',
     deploymentName: 'platform-primary',
     version: node['edgeCloudVersion'],
     services: svc_vars_primary,
     hostvols: hostvol_vars,
     configmaps: configmap_vars
   )
  only_if { node.attribute?(:redisServicePort) }
end

template '/home/ubuntu/k8s-deployment-secondary.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'secondary',
     deploymentName: 'platform-secondary',
     version: node['edgeCloudVersion'],
     services: svc_vars_secondary,
     hostvols: hostvol_vars,
     configmaps: configmap_vars
   )
  notifies :run, 'execute[delete-primary]', :immediately
  only_if { node.attribute?(:redisServicePort) }
end

# to affect a switchover, delete the primary deployment and re-create
execute('delete-primary') do
  action :nothing
  command 'kubectl delete -f /home/ubuntu/k8s-deployment-primary.yaml --kubeconfig=/home/ubuntu/.kube/config||true'
  returns 0
  notifies :run, 'execute[create-primary]', :immediately
  only_if { node.attribute?(:redisServicePort) }
end

execute('create-primary') do
  action :nothing
  command 'kubectl create -f /home/ubuntu/k8s-deployment-primary.yaml --kubeconfig=/home/ubuntu/.kube/config'
  returns 0
  notifies :run, 'execute[wait-primary]', :immediately
  only_if { node.attribute?(:redisServicePort) }
end

execute('wait-primary') do
  Chef::Log.info('Wait for primary platform pod to come up')
  action :nothing
  retries 30
  retry_delay 10
  command 'kubectl get pods -l app=platform-primary -l version=' + node['edgeCloudVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
  returns 0
  notifies :sleep, 'chef_sleep[sleep-after-primary]', :immediately
  only_if { node.attribute?(:redisServicePort) }
end

chef_sleep('sleep-after-primary') do
  seconds      60
  action       :nothing
  notifies :run, 'execute[delete-secondary]', :immediately
end

execute('delete-secondary') do
  action :nothing
  command 'kubectl delete -f /home/ubuntu/k8s-deployment-secondary.yaml --kubeconfig=/home/ubuntu/.kube/config||true'
  returns 0
  notifies :run, 'execute[create-secondary]', :immediately
  only_if { node.attribute?(:redisServicePort) }
end

execute('create-secondary') do
  action :nothing
  command 'kubectl create -f /home/ubuntu/k8s-deployment-secondary.yaml --kubeconfig=/home/ubuntu/.kube/config'
  returns 0
  notifies :run, 'execute[wait-secondary]', :immediately
  only_if { node.attribute?(:redisServicePort) }
end

execute('wait-secondary') do
  Chef::Log.info('Wait for seconday platform pod to come up')
  action :nothing
  retries 30
  retry_delay 10 
  command 'kubectl get pods -l app=platform-secondary -l version=' + node['edgeCloudVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
  returns 0
  only_if { node.attribute?(:redisServicePort) }
end
