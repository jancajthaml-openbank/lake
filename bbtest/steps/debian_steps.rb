step "lake contains following services" do |packages|
  items = packages.split("\n").map(&:strip).reject(&:empty?)

  containers = %x(docker ps -a -f name=lake | awk '{ print $1,$2 }' | grep #{ENV.fetch("VERSION", "latest")} | awk '{print $1 }' 2>/dev/null)
  expect($?).to be_success

  ids = containers.split("\n").map(&:strip).reject(&:empty?)
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually(timeout: 3) {
      units = %x(docker exec #{id} systemctl list-unit-files --type=service | awk '{ print $1 }' | grep .service | cat)
      units = ($? == 0 ? units.split("\n").map(&:strip).reject(&:empty?) : [])
      items.each { |item|
        expect(units).to include(item)
      }
    }
  }
end
