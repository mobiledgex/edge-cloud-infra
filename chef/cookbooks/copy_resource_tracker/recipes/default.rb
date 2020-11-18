if node.name.include? "mex-docker-vm"
  remote_file '/usr/local/bin/resource-tracker' do
#    source 'https://lev:APBLMkDUcg8Qinvm5dChcxMzAJ7@artifactory.mobiledgex.net/artifactory/repo-levdev/resource-tracker'
    source 'https://drive.google.com/uc?id=12pXTTZo9DSHsMPp7zhI0dE9xKqPHCH0o&export=download'
    checksum 'c594965648e20a2339d6f33d236b4e3e22b2be6916cceb1b0f338c74378c03da' #3233faadb0a371dc1cc787a08ab4c224
    mode '0771' # executable
    action :create
end 