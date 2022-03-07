default['upgrade_mobiledgex_package']['cert_validation'] = true

if node.exist?('upgrade_mobiledgex_package', 'version')
  curSemver = Gem::Version.new(node['upgrade_mobiledgex_package']['version'])
  certValSemver = Gem::Version.new("4.7.0")
  if curSemver < certValSemver
    default['upgrade_mobiledgex_package']['cert_validation'] = false
  end
end
