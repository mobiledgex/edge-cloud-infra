unified_mode true
resource_name :mobiledgex_apt_repository
provides :mobiledgex_apt_repository

property :cert_validation, [TrueClass, FalseClass], default: true

property :main_repo_url, String, name_property: true
property :main_repo_distribution, String, default: "bionic"
property :main_repo_components, Array, default: ["main"]
property :main_repo_username, String, default: "apt"
property :main_repo_password, String, default: "mobiledgex"

property :artifactory_repo_url, String, default: "https://artifactory.mobiledgex.net/artifactory/packages"
property :artifactory_repo_distribution, String, default: "cirrus"
property :artifactory_repo_components, Array, default: ["main"]
property :artifactory_repo_username, String, default: "apt"
property :artifactory_repo_password, String, default: "mobiledgex"

# Set up the apt repository
action :setup do

    # Set up apt cert validation
    file '/etc/apt/apt.conf.d/10cert-validation' do
        content "Acquire::https::Verify-Peer \"#{new_resource.cert_validation}\";\n"
        action :create
    end

    # Make sure the source list is empty
    file "/etc/apt/sources.list" do
        content ""
    end

    # Make sure the apt sources directory is present
    directory "/etc/apt/sources.list.d" do
        owner   "root"
        group   "root"
        mode    "0755"
        action  :create
    end

    # Set up the main apt repository
    apt_repository new_resource.main_repo_distribution do
        uri             new_resource.main_repo_url
        distribution    new_resource.main_repo_distribution
        components      new_resource.main_repo_components
    end

    # Set up the artifactory apt repository
    apt_repository new_resource.artifactory_repo_distribution do
        uri             new_resource.artifactory_repo_url
        distribution    new_resource.artifactory_repo_distribution
        components      new_resource.artifactory_repo_components
    end

    # Set up credentials for apt repositories
    template '/etc/apt/auth.conf.d/mobiledgex.net.conf' do
        source  "apt-auth.erb"
        owner   "root"
        group   "root"
        mode    "0400"
        variables(
            repos: [
                { :url      => new_resource.main_repo_url,
                  :user     => new_resource.main_repo_username,
                  :password => new_resource.main_repo_password,
                },
                { :url      => new_resource.artifactory_repo_url,
                  :user     => new_resource.artifactory_repo_username,
                  :password => new_resource.artifactory_repo_password,
                },
            ],
        )
    end

end
