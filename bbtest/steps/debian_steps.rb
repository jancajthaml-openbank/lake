step "systemctl contains following" do |packages|
  items = packages.split("\n").map(&:strip).reject(&:empty?)

  containers = %x(docker ps -a -f name=lake | awk '{ print $1,$2 }' | grep #{ENV.fetch("VERSION", "latest")} | awk '{print $1 }' 2>/dev/null)
  expect($?).to be_success

  ids = containers.split("\n").map(&:strip).reject(&:empty?)
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually(timeout: 3) {
      units = %x(docker exec #{id} systemctl list-unit-files --type=service | grep .service | awk '{ print $1 }')
      units = ($? == 0 ? units.split("\n").map(&:strip).reject(&:empty?) : [])
      expect(units).to include(*items)
    }
  }
end
