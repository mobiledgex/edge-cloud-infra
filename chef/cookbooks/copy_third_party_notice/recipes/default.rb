remote_file '/etc/THIRD_PARTY_NOTICE.txt' do
  source 'https://apt:AP2XYr1wBzePUAiKENupjzzB9ki@artifactory.mobiledgex.net/artifactory/downloads/THIRD_PARTY_NOTICE.txt'
  mode '0444'
  action :create
end
