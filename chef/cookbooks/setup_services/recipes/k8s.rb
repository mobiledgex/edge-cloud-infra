execute('Wait for K8s cluster to come up') do
  Chef::Log.info("Wait for K8s cluster to come up, there should be #{node['k8sNodeCount']} number of nodes")
  action 'run'
  retries 30
  retry_delay 6
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
template '/home/ubuntu/k8s-deployment-redis-master.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'master',
     services: {
          redis: {
              image: node['redisImage'] + ':' + node['redisVersion'],
              serviceName: node['redisServiceName'],
              port: node['redisPort'],
          },
     },
     hostvols: {},
     configmaps: {}
   )
  only_if node['redisPort']
end

svc_vars = get_services_vars
hostvol_vars = get_hostvols_vars
configmap_vars = get_configmap_vars

# start processes only any node if no HA enabled
template '/home/ubuntu/k8s-deployment.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'simplex',
     services: svc_vars,
     hostvols: hostvol_vars,
     configmaps: configmap_vars
   )
  not_if node['redisPort']
end

# start primary platform if HA enabled
template '/home/ubuntu/k8s-deployment-primary.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'primary',
     services: svc_vars,
     hostvols: hostvol_vars,
     configmaps: configmap_vars
   )
  only_if node['redisPort']
end

# start secondary platform if HA enabled
template '/home/ubuntu/k8s-deployment-secondary.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'secondary',
     services: svc_vars,
     hostvols: hostvol_vars,
     configmaps: configmap_vars
   )
  only_if node['redisPort']
end

execute('Setup kube pods primary') do
  action 'run'
  command 'kubectl apply -f /home/ubuntu/k8s-deployment-primary.yaml --kubeconfig=/home/ubuntu/.kube/config'
  returns 0
end

execute('Wait for primary deployment to come up') do
  Chef::Log.info("Wait for K8s cluster to come up, there should be #{node['k8sNodeCount']} number of nodes")
  action 'run'
  retries 30
  retry_delay 6
  command "kubectl get nodes --kubeconfig=/home/ubuntu/.kube/config| grep ' Ready' | wc -l | grep -w #{node['k8sNodeCount']}"
  returns 0
end
