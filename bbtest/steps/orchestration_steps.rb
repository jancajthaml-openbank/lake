step "lake is restarted" do ||
  ids = %x(systemctl -a -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("lake-")
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
    out = %x(systemctl show -p SubState lake-relay 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "lake is reconfigured with" do |configuration|
  params = Hash[configuration.split("\n").map(&:strip).reject(&:empty?).map { |el| el.split '=' }]
  defaults = {
    "LOG_LEVEL" => "DEBUG",
    "PORT_PULL" => "5562",
    "PORT_PUB" => "5561",
    "METRICS_REFRESHRATE" => "1h",
    "METRICS_OUTPUT" => "/reports/bbtest",
    "METRICS_CONTINOUS" => "true",
  }

  config = Array[defaults.merge(params).map { |k,v| "LAKE_#{k}=#{v}"} ]
  config = config.join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{config}' > /etc/init/lake.conf)

  %x(systemctl restart lake-relay 2>&1)

  eventually() {
    out = %x(systemctl show -p SubState lake-relay 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end
