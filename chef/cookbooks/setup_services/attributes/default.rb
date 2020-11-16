# controller notify address
ctrlNotifyAddrs = node.normal['crmserver']['args']['notifyAddrs']
ctrlAddr = ctrlNotifyAddrs.split(':')[0]
# access API address
accessApiAddr = ctrlAddr + ':41001'

services = ['crmserver', 'shepherd']
services.each { |service|
  # Set ES_SERVER_URLS environment variable
  deployTag = node.normal[service]['args']['deploymentTag']
  unless node.normal[service]['env'].any? { |s| s.include?('ES_SERVER_URLS') }
    Chef::Log.info("Setting ES_SERVER_URLS env var for #{service} belonging to #{deployTag} setup...")
    if deployTag == "main"
      node.normal[service]['env'].append("ES_SERVER_URLS=https://events.es.mobiledgex.net/")
    else
      node.normal[service]['env'].append("ES_SERVER_URLS=https://events-#{deployTag}.es.mobiledgex.net/")
    end
  end

  # Set accessApiAddr attribute for all the services
  unless node.normal[service]['args'].any? { |s| s.include?('accessApiAddr') }
    Chef::Log.info("Setting accessApiAddr attribute var for #{service} to #{accessApiAddr}...")
    node.normal[service]['args']['accessApiAddr'] = accessApiAddr
  end
}
