# -*- mode: ruby -*-
# vi: set ft=ruby :

require 'fileutils'

Vagrant.require_version ">= 2.1.0"
VAGRANTFILE_API_VERSION = "2"

$update_channel = "alpha"
$shared_folders = {
  "./" => "/lastbackend"
}

# default hub settings
$hub_count = 1
$hub_vm_memory = 512
# default builder settings
$builder_count = 1
$builder_vm_memory = 512
# default postgres settings
$postgres_vm_memory = 512
# default registry settings
$registry_vm_memory = 512


CONFIG = File.join(File.dirname(__FILE__), "./contrib/vagrant/config.rb")
if File.exist?(CONFIG)
  require CONFIG
end

if $builder_vm_memory < 1024
  puts "Builders should have at least 1024 MB of memory"
end

HUB_IP="10.3.0.1"
POSTGRES_IP="10.3.0.2"
REGISTRY_IP="10.3.0.3"

HUB_CLOUD_CONFIG_PATH = File.expand_path("./contrib/vagrant/hub-install.sh")
BUILDER_CLOUD_CONFIG_PATH = File.expand_path("./contrib/vagrant/builder-install.sh")
POSTGRES_CLOUD_CONFIG_PATH = File.expand_path("./contrib/vagrant/postgres-install.sh")
REGISTRY_CLOUD_CONFIG_PATH = File.expand_path("./contrib/vagrant/registry-install.sh")

def hubIP(num)
  return "172.17.4.#{num+100}"
end

def builderIP(num)
  return "172.17.4.#{num+200}"
end

hubIPs = [*1..$hub_count].map{ |i| hubIP(i) } <<  HUB_IP
builderIPs = [*1..$hub_count].map{ |i| builderIP(i) } <<  HUB_IP

# Generate root CA
# ISSUE: https://github.com/hashicorp/vagrant/issues/7747
system("mkdir -p ssl && ./hack/ssl/init-ssl-ca ssl") or abort ("failed generating SSL artifacts")

# Generate admin key/cert
system("./hack/ssl/init-ssl ssl admin lb-admin") or abort("failed generating admin SSL artifacts")

def provisionMachineSSL(machine,certBaseName,cn,ipAddrs)
  tarFile = "ssl/#{cn}.tar"
  ipString = ipAddrs.map.with_index { |ip, i| "IP.#{i+1}=#{ip}"}.join(",")

  system("./hack/ssl/init-ssl ssl #{certBaseName} #{cn} #{ipString}") or abort("failed generating #{cn} SSL artifacts")
  machine.vm.provision :file, :source => tarFile, :destination => "/tmp/ssl.tar"
  machine.vm.provision :shell, :inline => "mkdir -p /etc/lastbackend/ssl && tar -C /etc/lastbackend/ssl -xf /tmp/ssl.tar", :privileged => true
end

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  # always use Vagrant's insecure key
  config.ssh.insert_key = false

  config.vm.box = "coreos-%s" % $update_channel
  config.vm.box_version = ">= 1828.0.0"
  config.vm.box_url = "http://%s.release.core-os.net/amd64-usr/current/coreos_production_vagrant.json" % $update_channel

  config.vm.provider :virtualbox do |v|
    v.cpus = 1
    v.gui = false

    # On VirtualBox, we don't have guest additions or a functional vboxsf
    # in CoreOS, so tell Vagrant that so it can be smarter.
    v.check_guest_additions = false
    v.functional_vboxsf     = false
  end

  # plugin conflict
  if Vagrant.has_plugin?("vagrant-vbguest") then
    config.vbguest.auto_update = false
  end

  $shared_folders.each_with_index do |(host_folder, guest_folder), i|
    config.vm.synced_folder host_folder.to_s, guest_folder.to_s, id: "core-share%02d" % i, nfs: true, mount_options: ['nolock,vers=3,udp']
  end

  #config.vm.provision "docker" do |d|
  #  d.build_image "-t lastbackend/registry -f /lastbackend/images/registry/Dockerfile /lastbackend"
  #end

  (1..$hub_count).each do |i|
    config.vm.define vm_name = "h%d" % i do |hub|

      hub.vm.hostname = vm_name
      hub.vm.provider :virtualbox do |vb|
        vb.memory = $hub_vm_memory
      end

      hubIP = hubIP(i)
      hub.vm.network :private_network, ip: hubIP

      # Each hub gets the same cert
      provisionMachineSSL(hub,"hub","lb-hub-#{hubIP}",[hubIP])
      provisionMachineSSL(hub,"builder","lb-builder-#{hubIP}",builderIPs)

      hub.vm.provision :file, :source => HUB_CLOUD_CONFIG_PATH, :destination => "/tmp/vagrantfile-user-data"
      hub.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true

    end
  end

  (1..$builder_count).each do |i|
    config.vm.define vm_name = "b%d" % i do |builder|

      builder.vm.hostname = vm_name
      builder.vm.provider :virtualbox do |vb|
        vb.memory = $builder_vm_memory
      end

      builderIP = builderIP(i)
      builder.vm.network :private_network, ip: builderIP

      provisionMachineSSL(builder,"builder","lb-builder-#{builderIP}",[builderIP])
      provisionMachineSSL(builder,"hub","lb-hub-#{builderIP}",hubIPs)

      builder.vm.provision :file, :source => BUILDER_CLOUD_CONFIG_PATH, :destination => "/tmp/vagrantfile-user-data"
      builder.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true

    end
  end

  config.vm.define vm_name = "pg" do |postgres|

    postgres.vm.hostname = vm_name
    postgres.vm.provider :virtualbox do |vb|
      vb.memory = $postgres_vm_memory
    end

    postgres.vm.provision "docker" do |d|
      d.pull_images "postgres"
    end

    postgres.vm.network :private_network, ip: POSTGRES_IP
    postgres.vm.provision :file, :source => POSTGRES_CLOUD_CONFIG_PATH, :destination => "/tmp/vagrantfile-user-data"
    postgres.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true

  end

  config.vm.define vm_name = "rg" do |registry|

    registry.vm.hostname = vm_name
    registry.vm.provider :virtualbox do |vb|
      vb.memory = $registry_vm_memory
    end

    provisionMachineSSL(registry,"registry","lb-registry-#{REGISTRY_IP}",[REGISTRY_IP])

    registry.vm.provision "docker" do |d|
      d.pull_images "registry:2"
    end

    registry.vm.network :private_network, ip: REGISTRY_IP
    registry.vm.provision :file, :source => REGISTRY_CLOUD_CONFIG_PATH, :destination => "/tmp/vagrantfile-user-data"
    registry.vm.provision :shell, :inline => "mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/", :privileged => true

  end

end