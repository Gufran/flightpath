Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/eoan64"
  config.vm.box_check_update = false

  config.vm.network "forwarded_port", guest: 4646, host: 4646
  config.vm.network "forwarded_port", guest: 8500, host: 8500
  config.vm.network "forwarded_port", guest: 9292, host: 9292
  config.vm.network "forwarded_port", guest: 9998, host: 9998
  config.vm.network "forwarded_port", guest: 9901, host: 9901

  config.vm.provider "virtualbox" do |vb|
    vb.name = "flightpath"
    vb.linked_clone = true
    vb.memory = 2048
  end

  config.vm.provision "bootstrap", type: "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y zip ca-certificates curl apt-transport-https gnupg-agent software-properties-common

    curl -s https://releases.hashicorp.com/nomad/0.10.5/nomad_0.10.5_linux_amd64.zip -o /tmp/nomad.zip
    unzip /tmp/nomad.zip -d /usr/bin

    curl -s -L https://github.com/containernetworking/plugins/releases/download/v0.8.4/cni-plugins-linux-amd64-v0.8.4.tgz -o /tmp/cni-plugins.tgz
    mkdir -p /opt/cni/bin
    tar -C /opt/cni/bin -xzf /tmp/cni-plugins.tgz

    curl -s https://releases.hashicorp.com/consul/1.7.2/consul_1.7.2_linux_amd64.zip -o /tmp/consul.zip
    unzip /tmp/consul.zip -d /usr/bin

    useradd envoy -U -M
    curl -s -L https://getenvoy.io/cli | bash -s -- -b /usr/local/bin
    /usr/local/bin/getenvoy fetch standard:1.13.1
    mv $HOME/.getenvoy/builds/standard/1.13.1/linux_glibc/bin/envoy /usr/bin/envoy
    mkdir /var/log/envoy
    cp /vagrant/internal/testing/config/envoy.yaml /etc/envoy.yaml
    chown envoy:envoy /var/log/envoy /etc/envoy.yaml

    setcap cap_net_bind_service,cap_net_admin=+ep /usr/bin/envoy

    chmod +x /usr/bin/consul /usr/bin/nomad /usr/bin/envoy

    curl -s https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz -o /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/bash.bashrc

    rm -rf /tmp/nomad.zip /tmp/consul.zip /usr/local/bin/genenvoy $HOME/.getenvoy /tmp/go.tar.gz /tmp/cni-plugins.tgz

    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io
    usermod -aG docker vagrant

    cp /vagrant/internal/testing/config/consul.service /usr/lib/systemd/system/consul.service
    cp /vagrant/internal/testing/config/envoy.service /usr/lib/systemd/system/envoy.service
    cp /vagrant/internal/testing/config/flightpath.service /usr/lib/systemd/system/flightpath.service
    cp /vagrant/internal/testing/config/nomad.service /usr/lib/systemd/system/nomad.service

    systemctl enable consul envoy nomad
    systemctl start consul envoy nomad
  SHELL

  config.vm.provision "flightpath", type: "shell", after: "bootstrap", inline: <<-SHELL
    export PATH=$PATH:/usr/local/go/bin

    cd /vagrant
    ./build.sh native
    mv flightpath /usr/bin/flightpath

    systemctl enable flightpath
    systemctl start flightpath
  SHELL

  config.vm.provision "jobs", type: "shell", after: "flightpath", inline: <<-SHELL
    sleep 5
    nomad job run /vagrant/internal/testing/jobspec/with-connect.nomad
    nomad job run /vagrant/internal/testing/jobspec/without-connect.nomad
  SHELL
end
