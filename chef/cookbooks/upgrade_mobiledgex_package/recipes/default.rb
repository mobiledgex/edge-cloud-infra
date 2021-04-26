if node.normal['tags'].include?('vmtype/rootlb')
  # Check installed package `dpkg -s mobiledgex | grep Version | awk -F ": " '{print $2}'`
  curTag = shell_out("dpkg -s mobiledgex | grep Version | awk -F \": \" '{printf $2}'").stdout
  unless node['mobiledgeXPackageVersion'] == curTag
    # First delete old pin file if it exists
    file '/etc/apt/preferences.d/99mobiledgex' do
      action :delete
      only_if { File.exist? '/etc/apt/preferences.d/99mobiledgex' }
    end
    # Pin mobiledge package version
    apt_preference 'mobiledgex' do
      pin "version #{node['mobiledgeXPackageVersion']}"
      pin_priority '1001'
    end
    apt_package 'mobiledgex' do
      action :upgrade
    end
  end
end
