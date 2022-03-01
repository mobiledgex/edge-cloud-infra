unified_mode true
resource_name :platform_k8s
provides :platform_k8s

action :prep_cluster do
  execute("wait-k8s-cluster, looking for #{node['k8sNodeCount']} nodes") do
    action :run
    retries 30
    retry_delay 10
    command "kubectl get nodes --kubeconfig=/home/ubuntu/.kube/config| grep ' Ready' | wc -l | grep -w #{node['k8sNodeCount']}"
    returns 0
  end

  execute('Setup docker registry secrets') do
    action :run
    retries 2
    retry_delay 2
    regsecrets = data_bag_item('mexsecrets', 'docker_registry')
    Chef::Log.info("Create secret to access #{node['edgeCloudImage']} as user #{regsecrets['mex_docker_username']}")
    command "kubectl create secret docker-registry mexreg-secret --docker-server=#{node['edgeCloudImage']} --docker-username=#{regsecrets['mex_docker_username']} --docker-password=#{regsecrets['mex_docker_password']} --docker-email=mobiledgex.ops@mobiledgex.com --kubeconfig=/home/ubuntu/.kube/config"
    returns 0
    ignore_failure true
  end

  execute('Remove master taint') do
    action :run
    retries 2
    retry_delay 2
    command 'kubectl taint nodes -l node-role.kubernetes.io/master node-role.kubernetes.io/master:NoSchedule- --kubeconfig=/home/ubuntu/.kube/config'
    returns 0
    ignore_failure true
  end

#  execute('Assign k8s node labels for master') do
#    action :run
#    retries 2
#    retry_delay 2
#    Chef::Log.info("Setting label platform-cluster-master for #{node['platform-cluster-master']} ")
#    command "kubectl label nodes #{node['platform-cluster-master']} harole=master --kubeconfig=/home/ubuntu/.kube/config"
#    returns 0
#    ignore_failure true
#  end

#  execute('Assign k8s node labels for primary') do
#    action :run
#    retries 2
#    retry_delay 2
#    Chef::Log.info("Setting label platform-cluster-primary-node for #{node['platform-cluster-primary-node']} ")
#    command "kubectl label nodes #{node['platform-cluster-primary-node']} harole=primary --kubeconfig=/home/ubuntu/.kube/config"
#    returns 0
#    ignore_failure true
#  end

#  execute('Assign k8s node labels for secondary') do
#    action :run
#    retries 2
#    retry_delay 2
#    Chef::Log.info("Setting label platform-cluster-secondary-node for #{node['platform-cluster-secondary-node']} ")
#    command "kubectl label nodes #{node['platform-cluster-secondary-node']} harole=secondary --kubeconfig=/home/ubuntu/.kube/config"
#    returns 0
#    ignore_failure true
#  end
end # prep-cluster

action :setup_redis do
  execute('Setup redis deployment') do
    Chef::Log.info("Setting up redis")
    action :run
    command 'kubectl apply -f /home/ubuntu/k8s-deployment-redis.yaml --kubeconfig=/home/ubuntu/.kube/config'
    returns 0
  end

  execute('Wait for redis deployment to come up') do
    Chef::Log.info('Wait for redis deployment to come up')
    action :run
    retries 20
    retry_delay 6
    command 'kubectl get pods -l app=redis -l version=' + node['redisVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
    returns 0
    only_if { node.attribute?(:redisServicePort) }
  end
end # setup-redis

action :deploy_simplex_platform do
  # simplex needs prom too
  Chef::Log.info('Deploy prometheus platform simplex')
  cookbook_file 'home/ubuntu/prometheus.yml' do
    source 'prometheus.yml'
    owner 'ubuntu'
    group 'ubuntu'
    mode '0644'
    action :create_if_missing
  end

  # PV and PVCs for accesskey and cache mounts into crm and shepherd containers
  cookbook_file 'home/ubuntu/accesskey_vol.yml'  do
    source 'accesskey_vol.yml'
    owner 'ubuntu'
    group 'ubuntu'
    mode '0664'
    action :create_if_missing
  end

  execute('apply-accesskey-vol') do
    Chef::Log.info('create accesskey vol peristent volume')
    action :run
    command 'kubectl apply -f /home/ubuntu/accesskey_vol.yml --kubeconfig=/home/ubuntu/.kube/config'
    retries 2
    retry_delay 2
    returns 0
  end

  cookbook_file 'home/ubuntu/accesskey_vol_claim.yml' do
    source 'accesskey_vol_claim.yml'
    owner 'ubuntu'
    group 'ubuntu'
    mode '0664'
    action :create_if_missing
  end

  execute('apply-accesskey-vol-claim') do
    Chef::Log.info('create accesskey vol peristent volume claim')
    action :run
    command 'kubectl apply -f /home/ubuntu/accesskey_vol_claim.yml --kubeconfig=/home/ubuntu/.kube/config'
    retries 2
    retry_delay 2
    returns 0
  end

  cookbook_file 'home/ubuntu/cache_vol.yml' do
    source 'cache_vol.yml'
    owner 'ubuntu'
    group 'ubuntu'
    mode '0664'
    action :create_if_missing
  end

  execute('apply-cache-vol') do
    Chef::Log.info('create cache vol peristent volume')
    action :run
    command 'kubectl apply -f /home/ubuntu/cache_vol.yml --kubeconfig=/home/ubuntu/.kube/config'
    retries 2
    retry_delay 2
    returns 0
  end

  cookbook_file 'home/ubuntu/cache_vol_claim.yml' do
    source 'cache_vol_claim.yml'
    owner 'ubuntu'
    group 'ubuntu'
    mode '0664'
    action :create_if_missing
  end

  execute('apply-cache-vol-claim') do
    Chef::Log.info('create cache vol peristent volume claim')
    action :run
    command 'kubectl apply -f /home/ubuntu/cache_vol_claim.yml --kubeconfig=/home/ubuntu/.kube/config'
    retries 2
    retry_delay 2
    returns 0
  end

  execute('create-prometheus-configmap') do
    Chef::Log.info('create prometheus configmap')
    action :run
    command 'kubectl create configmap prom-cm --from-file prometheus.yml=/home/ubuntu/prometheus.yml --kubeconfig=/home/ubuntu/.kube/config'
    retries 2
    retry_delay 2
    returns 0
    ignore_failure true
  end

  # if this works, maybe add cloudletorg to these args?
  execute('copy-kubeconfig-default-cluster') do
    cloudletorg = #node['crmserver']['args']['physicalname']
    filename = "defaultcloudlet." + cloudletorg + ".kubeconfig"
    Chef::Log.info('create file #filename')    
    action :run
    command 'cp /home/ubuntu/.kube/config #filename'
    returns 0
  end
  
  execute('Setup simplex deployment') do
    Chef::Log.info("apply k8s-deployment.yaml")
    action :run
    command 'kubectl apply -f /home/ubuntu/k8s-deployment.yaml --kubeconfig=/home/ubuntu/.kube/config'
    returns 0
  end

  execute('Wait for simplex platform pod to come up') do
    Chef::Log.info('Wait for simplex platform pod to come up')
    action :run
    retries 30
    retry_delay 10
    command 'kubectl get pods -l app=platform-simplex -l version=' + node['edgeCloudVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
    returns 0
  end

end # deploy-simplex-platform

#action :deploy_ha_platform do
#  cookbook_file 'home/ubuntu/prometheus.yml' do
#    source 'prometheus.yml'
#    owner 'ubuntu'
#    group 'ubuntu'
#    mode '0644'
#    action :create_if_missing
#  end
#
#  execute('create-prometheus-configmap') do
#    Chef::Log.info('create prometheus configmap')
#    action :run
#    command 'kubectl create configmap prom-cm --from-file prometheus.yml=/home/ubuntu/prometheus.yml --kubeconfig=/home/ubuntu/.kube/config'
#    retries 2
#    retry_delay 2
#    returns 0
#    ignore_failure true
#  end
#
#  # to affect a switchover, delete the primary deployment and re-create
#  execute('delete-primary') do
#    action :run
#    command 'kubectl delete -f /home/ubuntu/k8s-deployment-primary.yaml --kubeconfig=/home/ubuntu/.kube/config||true'
#    returns 0
#  end
#
#  execute('create-primary') do
#    action :run
#    command 'kubectl create -f /home/ubuntu/k8s-deployment-primary.yaml --kubeconfig=/home/ubuntu/.kube/config'
#    returns 0
#  end
#
#  execute('wait-primary') do
#    Chef::Log.info('Wait for primary platform pod to come up')
#    action :run
#    retries 30
#    retry_delay 10
#    command 'kubectl get pods -l app=platform-primary -l version=' + node['edgeCloudVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
#    returns 0
#  end
#
#  chef_sleep('sleep-after-primary') do
#    seconds      30
#    action       :sleep
#  end
#
#  execute('delete-secondary') do
#    action :run
#    command 'kubectl delete -f /home/ubuntu/k8s-deployment-secondary.yaml --kubeconfig=/home/ubuntu/.kube/config||true'
#    returns 0
#  end
#
#  execute('create-secondary') do
#    action :run
#    command 'kubectl create -f /home/ubuntu/k8s-deployment-secondary.yaml --kubeconfig=/home/ubuntu/.kube/config'
#    returns 0
#  end
#
#  execute('wait-secondary') do
#    Chef::Log.info('Wait for seconday platform pod to come up')
#    action :run
#    retries 30
#    retry_delay 10
#    command 'kubectl get pods -l app=platform-secondary -l version=' + node['edgeCloudVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'
#    returns 0
#  end
#end # deploy-ha-platform
