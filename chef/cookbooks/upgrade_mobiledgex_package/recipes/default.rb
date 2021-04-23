if node.role?("mex-agent-node")
  Chef::Log.info("Checking mobiledgex package version on: #{node.name}")
  # Check installed package `dpkg -s mobiledgex | grep Version | awk -F ": " '{print $2}'`
  curTag = shell_out("dpkg -s mobiledgex | grep Version | awk -F \": \" '{print $2}'").stdout
  Chef::Log.info("installed package tag: #{curTag} current package: #{node['mexInfraTag']}") 
  unless node['mexInfraTag'] == curTag
    Chef::Log.info("Creating a pinning file....")
    apt_preference 'mobiledgex' do
      pin "#{node['mexInfraTag']}"
      priority '1001'
    end
    Chef::Log.info("Installing new mobiledgeX package....")
    apt_package 'mobiledgex'
    newTag = shell_out("dpkg -s mobiledgex | grep Version | awk -F \": \" '{print $2}'").stdout
    Chef::Log.info("New installed package tag: #{curTag}")
  end
end
