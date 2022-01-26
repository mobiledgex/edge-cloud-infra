svc_vars_primary = get_services_vars('primary')
svc_vars_secondary = get_services_vars('secondary')
hostvol_vars = get_hostvols_vars
configmap_vars = get_configmap_vars

platform_k8s('prep platform cluster') do
  action :prep_cluster
  # setting up secondary role is the last step in prep, so if this exists do not run
  not_if 'kubectl get nodes --show-labels --kubeconfig=/home/ubuntu/.kube/config|grep harole=secondary'
end

platform_k8s('setup redis') do
  action :setup_redis
  only_if { node.attribute?(:redisServicePort) }
  not_if " 'kubectl get pods -l app=redis -l version=' + node['redisVersion'] + ' --kubeconfig=/home/ubuntu/.kube/config| grep Running'"
end

# If redis is not specified, create the simplex manifest. If modified, trigger simplex deployment
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
  notifies :run, :deploy_simplex_platform, immediately
end

# If redis is specified, create the primary and secondary manifests. If modified, trigger H/A deployment
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
  notifies :run, :deploy_ha_platform, :immediately
  only_if { node.attribute?(:redisServicePort) }
end
