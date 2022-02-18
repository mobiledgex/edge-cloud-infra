if node.attribute?('teleport_token')

    # Install teleport
    remote_file '/usr/local/bin/teleport' do
        source 'https://apt:AP2XYr1wBzePUAiKENupjzzB9ki@artifactory.mobiledgex.net/artifactory/downloads/teleport/v8.2.0/teleport'
        checksum 'ed2e0d6282597a1aa4d8963a2bd3ed9a270d6f177108087eb5b962d249dbdfa5'
        mode '0755'
        action :create
    end


    begin

        # Write the teleport initial token to a file
        file '/etc/teleport.token' do
            content "#{node.attribute['teleport_token']}"
            mode '0400'
            owner 'root'
            group 'root'

            notifies :restart, 'systemd_unit[teleport.service]', :delayed
        end

        # Get teleport operator label from cloudlet org
        cloudletorg = node.normal['tags'].find {|t| t.start_with?('cloudletorg')}.split("/")[1].downcase

        # Set up systemd service
        systemd_unit 'teleport.service' do
            content({
                Unit: {
                    Description: 'Teleport SSH Service',
                    After: 'network.target',
                },
                Service: {
                    Type: 'simple',
                    Restart: 'on-failure',
                    ExecStart: "/usr/local/bin/teleport start --roles=node --labels=env=ops,operator=#{cloudletorg} --token=/etc/teleport.token --auth-server=teleport.mobiledgex.net:443",
                    ExecReload: '/bin/kill -HUP $MAINPID',
                    PIDFile: '/run/teleport.pid',
                },
                Install: {
                    WantedBy: 'multi-user.target',
                }
            })
            action [:create, :enable, :start]
        end

    rescue => error

        log 'error' do
            message "Caught error #{error}"
            level :error
        end

        # Remove systemd service
        systemd_unit 'teleport.service' do
            action [:stop, :disable, :delete]
        end

    end

else
    # Not a teleport node; clean up stuff

    # Delete teleport file
    file '/usr/local/bin/teleport' do
        action :delete
    end

end

systemd_unit 'teleport.service' do
    action :nothing
end
