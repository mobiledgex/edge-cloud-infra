edgeCloudVersion = node['edgeCloudVersion']
if node.attribute?("edgeCloudVersionOverride")
  unless node['edgeCloudVersionOverride'].empty?
    Chef::Log.info("Override edgeCloudVersion from #{node['edgeCloudVersion']} to #{node['edgeCloudVersionOverride']}")
    edgeCloudVersion = node['edgeCloudVersionOverride']
  end
end

Chef::Log.info("Using edgeCloudVersion: #{edgeCloudVersion}")

docker_registry "#{node['edgeCloudImage']}" do
  regsecrets = data_bag_item('mexsecrets', 'docker_registry')
  Chef::Log.info("Login to registry: #{node['edgeCloudImage']} as user #{regsecrets['mex_docker_username']}")
  username "#{regsecrets['mex_docker_username']}"
  password "#{regsecrets['mex_docker_password']}"
end

docker_image "#{node['edgeCloudImage']}" do
  Chef::Log.info("Pull edge cloud image #{node['edgeCloudImage']}:#{edgeCloudVersion}")
  action :pull
  tag "#{edgeCloudVersion}"
  notifies :run, 'execute[prune-old-images]', :delayed
end

docker_image "docker.mobiledgex.net/mobiledgex/mobiledgex_public/#{node['prometheusImage']}" do
  Chef::Log.info("Pull prometheus image #{node['prometheusImage']}:#{node['prometheusVersion']}")
  action :pull
  tag "#{node['prometheusVersion']}"
  notifies :run, 'execute[prune-old-images]', :delayed
end

directory '/root/accesskey' do
  owner 'root'
  group 'root'
  mode '0700'
  action :create
end

directory '/root/crm_cache' do
  owner 'root'
  group 'root'
  mode '0700'
  action :create
end

# create    - Creates the container but does not start it. Useful for Volume containers.
# start     - Starts the container. Useful for containers that run jobs.. command that exit.
# run       - The default action. Both :create and :start the container in one action. Redeploys the container on resource change.
# stop      - Stops the container.
# restart   - Stops and then starts the container.
# delete    - Deletes the container.
# redeploy  - Deletes and runs the container.
dockerContainerActions = ['create', 'start', 'run', 'stop', 'restart', 'delete', 'redeploy']
defaultContainerAction = 'run'
edgeCloudContainerAction = defaultContainerAction
if node.attribute?('edgeCloudContainerAction')
  if node['edgeCloudContainerAction'].empty?
      Chef::Log.info("Using default container action value: " + defaultContainerAction)
      edgeCloudContainerAction = defaultContainerAction
  else
    if dockerContainerActions.include?("#{node['edgeCloudContainerAction']}")
      Chef::Log.info("Setting action on edge cloud containers: #{node['edgeCloudContainerAction']}")
      edgeCloudContainerAction = node['edgeCloudContainerAction']
    else
      Chef::Log.info("Invalid container action #{node['edgeCloudContainerAction']}, valid actions are " + dockerContainerActions.join(","))
      Chef::Log.info("Using default container action value: " + defaultContainerAction)
      edgeCloudContainerAction = defaultContainerAction
    end
  end
else
  Chef::Log.info("Using default container action value: " + defaultContainerAction)
  edgeCloudContainerAction = defaultContainerAction
end

cmd = crmserver_cmd
crmserver_volumes = [
  '/var/tmp:/var/tmp',
  '/root/accesskey:/root/accesskey',
  '/root/crm_cache:/root/crm_cache'
]
if File.file? '/etc/mex-release'
  crmserver_volumes.append('/etc/mex-release:/etc/mex-release')
end
docker_container "crmserver" do
  Chef::Log.info("Performing action '#{edgeCloudContainerAction}' on crmserver container, cmd: #{cmd}")
  repo "#{node['edgeCloudImage']}"
  tag "#{edgeCloudVersion}"
  action "#{edgeCloudContainerAction}"
  network_mode 'host'
  restart_policy 'unless-stopped'
  env node['crmserver']['env']
  volumes crmserver_volumes
  command cmd
end

cmd = shepherd_cmd
docker_container "shepherd" do
  Chef::Log.info("Performing action '#{edgeCloudContainerAction}' on shepherd container, cmd: #{cmd}")
  repo "#{node['edgeCloudImage']}"
  tag "#{edgeCloudVersion}"
  action "#{edgeCloudContainerAction}"
  network_mode 'host'
  restart_policy 'unless-stopped'
  env node['shepherd']['env']
  volumes ['/tmp:/tmp', '/root/accesskey:/root/accesskey']
  command cmd
end

template '/tmp/prometheus.yml' do
  source 'prometheus.erb'
  variables(
    remote_write_addr: get_thanos_remote_write_addr()
  )
  mode '0644'
  action :create
  force_unlink true
  notifies :restart, 'docker_container[cloudletPrometheus]', :delayed
end

cmd = cloudlet_prometheus_cmd
docker_container "cloudletPrometheus" do
  Chef::Log.info("Performing action '#{edgeCloudContainerAction}' on cloudlet prometheus container, cmd: #{cmd}")
  repo "docker.mobiledgex.net/mobiledgex/mobiledgex_public/#{node['prometheusImage']}"
  tag "#{node['prometheusVersion']}"
  action "#{edgeCloudContainerAction}"
  network_mode 'host'
  restart_policy 'unless-stopped'
  if node['cloudletPrometheus'].key?("env")
    env node['cloudletPrometheus']['env']
  end
  if node['cloudletPrometheus'].key?("publish")
    port node['cloudletPrometheus']['publish']
  end
  if node['cloudletPrometheus'].key?("label")
    labels node['cloudletPrometheus']['label']
  end
  if node['cloudletPrometheus'].key?("volume")
    volumes node['cloudletPrometheus']['volume']
  end
  user "nobody"
  working_dir "/prometheus"
  command cmd
end

# Prune old docker images only when a new docker image is pulled
execute "prune-old-images" do
  command 'docker image prune -a --force --filter "until=24h"'
  action :nothing
end
