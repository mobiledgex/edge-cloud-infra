defaultPackageVertsion = '4.3.4'

if node.normal['tags'].include?('vmtype/rootlb')
  # Check installed package `dpkg -s mobiledgex | grep Version | awk -F ": " '{print $2}'`
  curTag = shell_out("dpkg -s mobiledgex | grep Version | awk -F \": \" '{print $2}'").stdout
  # set mexInfraTag attribute to default if it's not set
  unless node.attribute?('mexInfraTag')
    node.default['mexInfraTag'] = defaultPackageVertsion
  end
  unless node['mexInfraTag'] == curTag
    # First delete old pin file if it exists
    file '/etc/apt/preferences.d/99mobiledgex' do
      action :delete
      only_if { File.exist? '/etc/apt/preferences.d/99mobiledgex' }
    end
    # Pin mobiledge package version
    apt_preference 'mobiledgex' do
      pin "version #{node['mexInfraTag']}"
      pin_priority '1001'
    end
    apt_package 'mobiledgex' do
      action :upgrade
    end
    newTag = shell_out("dpkg -s mobiledgex | grep Version | awk -F \": \" '{print $2}'").stdout
  end
end
