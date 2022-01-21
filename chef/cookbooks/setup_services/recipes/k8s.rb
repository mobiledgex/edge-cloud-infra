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

crm_cmd = crmserver_cmd
shp_cmd = shepherd_cmd
prom_cmd = cloudlet_prometheus_cmd
template '/home/ubuntu/k8s-deployment-primary.yaml' do
  source 'k8s_service.erb'
  variables(
     harole: 'primary',
     services: {
       crmserver: { cmd: crm_cmd,
                    env: node['crmserver']['env'],
                    image: node['edgeCloudImage'] + ':' + node['edgeCloudVersion'],
                    volumeMounts: { accesskey_vol: { name: 'accesskey-vol', mountPath: '/accesskey' } },
       },
       shepherd: { cmd: shp_cmd,
                   env: node['shepherd']['env'],
                   image: node['edgeCloudImage'] + ':' + node['edgeCloudVersion'],
                   volumeMounts: { accesskey_vol: { name: 'accesskey-vol', mountPath: '/accesskey' } },
       },
       cloudletprometheus: { cmd: prom_cmd,
                             env: node['cloudletPrometheus']['env'],
                             image: 'docker.mobiledgex.net/mobiledgex/mobiledgex_public/' + node['prometheusImage'] + ':' + node['prometheusVersion'],
                             volumeMounts: { prom_vol: { name: 'prom-config', mountPath: '/etc/prometheus' } },
        },
     },
     hostvols: { accesskey_vol: { name: 'accesskey-vol', hostPath: '/root/accesskey' } },
     configmaps: { prom_config: { name: 'prom-config', configMap: 'prom-cm', key: 'prometheus.yml', path: 'prometheus.yml' } }
   )
end

execute('Setup kube pods') do
  action 'run'
  command 'kubectl apply -f /tmp/k8s-deployment.yaml --kubeconfig=/home/ubuntu/.kube/config'
  returns 0
end
