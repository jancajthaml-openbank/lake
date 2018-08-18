
step "systemctl contains following" do |packages|
  items = packages.split("\n").map(&:strip).reject(&:empty?)

  ids = Docker.get_lakes()
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually(timeout: 3) {
      units = %x(docker exec #{id} systemctl list-unit-files --type=service | grep .service | awk '{ print $1 }')
      units = ($? == 0 ? units.split("\n").map(&:strip).reject(&:empty?) : [])
      expect(units).to include(*items)
    }
  }
end

step ":operation package :package" do |operation, package|
  ids = Docker.get_lakes()
  expect(ids).not_to be_empty

  ids.each { |id|
    %x(docker exec #{id} systemctl #{operation} #{package} 2>&1)

    unless $? == 0
      err = %x(docker exec #{id} systemctl status #{package} 2>&1)
      raise "operation \"systemctl #{operation} #{package}\" returned error: #{err}"
    end
  }
end

step "package :package is running" do |package|
  ids = Docker.get_lakes()
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually(timeout: 10) {
      package_status = %x(docker exec #{id} systemctl show -p SubState #{package} 2>&1 | sed 's/SubState=//g')
      expect(package_status.strip).to eq("running")
    }
  }
end

step "package :package is not running" do |package|
  ids = Docker.get_lakes()
  expect(ids).not_to be_empty

  ids.each { |id|

    eventually(timeout: 10) {
      package_status = %x(docker exec #{id} systemctl show -p SubState #{package} 2>&1 | sed 's/SubState=//g')
      expect(package_status.strip).to eq("dead")
    }
  }
end
