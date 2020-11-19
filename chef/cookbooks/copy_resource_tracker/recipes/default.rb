if node.name.include? "mex-docker-vm"
  remote_file '/usr/local/bin/resource-tracker' do
    source 'https://rtifactory.mobiledgex.net/artifactory/packages/resource-tracker' # TODO - change to the actual link
    checksum '3233faadb0a371dc1cc787a08ab4c224'   # TODO - change to the checksum in artifactory
    mode '0771' # executable
    action :create
  end
end
