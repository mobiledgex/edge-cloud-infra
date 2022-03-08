mobiledgex_apt_repository node['upgrade_mobiledgex_package']['repo'] do
  action          :setup
  cert_validation node['upgrade_mobiledgex_package']['cert_validation']
end

mobiledgex_pkg node['mobiledgeXPackageVersion'] do
  action  :upgrade
end
