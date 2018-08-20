
step "systemctl contains following" do |packages|
  items = packages.split("\n").map(&:strip).reject(&:empty?)

  eventually() {
    items.each { |item|
      units = %x(systemctl list-units --type=service | grep #{item} | awk '{ print $1 }')
      units = units.split("\n").map(&:strip).reject(&:empty?)
      expect(units).not_to be_empty, "#{item} was not found"
    }
  }
end

step "systemctl does not contains following" do |packages|
  items = packages.split("\n").map(&:strip).reject(&:empty?)

  items.each { |item|
    units = %x(systemctl list-units --type=service | grep #{item} | awk '{ print $1 }')
    units = units.split("\n").map(&:strip).reject(&:empty?)
    expect(units).to be_empty, "#{item} was not found"
  }
end

step ":operation unit :unit" do |operation, unit|
  eventually(timeout: 5) {
    %x(systemctl #{operation} #{unit} 2>&1)
  }

  unless $? == 0
    err = %x(systemctl status #{unit} 2>&1)
    raise "operation \"systemctl #{operation} #{unit}\" returned error: #{err}"
  end
end

step "unit :unit is running" do |unit|
  eventually() {
    out = %x(systemctl show -p SubState #{unit} 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "unit :unit is not running" do |unit|
  eventually() {
    out = %x(systemctl show -p SubState #{unit} 2>&1 | sed 's/SubState=//g')
    expect(out.strip).not_to eq("running")
  }
end
