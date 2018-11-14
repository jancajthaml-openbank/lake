step "lake is restarted" do ||
  ids = %x(systemctl -a -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("lake")
  }.map { |x| x.chomp(".service") }

  expect(ids).not_to be_empty

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end

step "lake is running" do ||
  eventually() {
    out = %x(systemctl show -p SubState lake 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }

  eventually(timeout: 5) {
    ZMQHelper.lake_handshake()
  }
end

step "lake is running with following configuration" do |configuration|
  params = configuration.split("\n").map(&:strip).reject(&:empty?).join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{params}' > /etc/init/lake.conf)

  %x(systemctl restart lake 2>&1)

  eventually() {
    out = %x(systemctl show -p SubState lake 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }

  eventually(timeout: 5) {
    ZMQHelper.lake_handshake()
  }
end
