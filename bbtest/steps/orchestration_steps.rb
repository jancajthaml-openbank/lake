step "lake is restarted" do ||
  %x(systemctl restart lake-relay.service 2>&1)

  eventually() {
    out = %x(systemctl show -p SubState lake-relay.service 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "lake is running" do ||
  eventually() {
    out = %x(systemctl show -p SubState lake-relay.service 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "lake is reconfigured with" do |configuration|
  params = Hash[configuration.split("\n").map(&:strip).reject(&:empty?).map { |el| el.split '=' }]
  config = Array[UnitHelper.default_config.merge(params).map { |k,v| "LAKE_#{k}=#{v}"} ]
  config = config.join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{config}' > /etc/init/lake.conf)

  %x(systemctl restart lake-relay.service 2>&1)

  eventually() {
    out = %x(systemctl show -p SubState lake-relay.service 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end
