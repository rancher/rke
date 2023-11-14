def dockerInstall(vm)
  vm.provider "libvirt" do |v|
    v.memory = NODE_MEMORY + 1024
  end
  vm.provider "virtualbox" do |v|
    v.memory = NODE_MEMORY + 1024
  end
  box = vm.box.to_s
  if box.include?("ubuntu")
    vm.provision "shell", inline: "apt update; apt install -y docker.io"
  elsif box.include?("Leap")
    vm.provision "shell", inline: "zypper install -y docker apparmor-parser"
  elsif box.include?("microos")
    vm.provision "shell", inline: "transactional-update pkg install -y docker apparmor-parser"
    vm.provision 'docker-reload', type: 'reload', run: 'once'
    vm.provision "shell", inline: "systemctl enable --now docker"
  elsif box.include?("rocky8") || box.include?("rocky9")
    vm.provision "shell", inline: "dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo"
    vm.provision "shell", inline: "dnf install -y docker-ce"
  end
  vm.provision "shell", inline: "usermod -aG docker vagrant"
end
