step "no containers are running" do ||
  containers = %x(docker ps -a | awk '{ print $1,$2 }' | grep openbank/lake | awk '{print $1 }' 2>/dev/null)
  containers = ($? == 0 ? containers.split("\n") : []).map(&:strip).reject(&:empty?)

  containers.each { |id|
    eventually(timeout: 3) {
      %x(docker kill --signal="TERM" #{id} >/dev/null 2>&1)
      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("false")

      label = %x(docker inspect --format='{{.Name}}' #{id})
      label = ($? == 0 ? label.strip : id)

      %x(docker logs #{id} >/logs/#{label}.log 2>&1)
      %x(docker rm -f #{id} &>/dev/null || :)
    }
  }
end

step "container is started" do ||
  send "no containers are running"

  id = %x(docker run \
    -d \
    -h lake \
    -p 5561 \
    -p 5562 \
    --net=lake_default \
    --net-alias=lake \
    --name=lake \
  openbank/lake:#{ENV.fetch("VERSION", "latest")} 2>&1)
  expect($?).to be_success, id

  eventually(timeout: 3) {
    container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
    expect($?).to be_success
    expect(container_state.strip).to eq("true")
  }
end

step "container should be running" do ||
  containers = %x(docker ps -a | awk '{ print $1,$2 }' | grep openbank/lake | awk '{print $1 }' 2>/dev/null)
  expect($?).to be_success
  expect(containers).not_to be_empty

  containers.split("\n").map(&:strip).reject(&:empty?).each { |id|
    send ":container running state is :state", id, true
  }
end

step ":container running state is :state" do |container, state|
  eventually(timeout: 3) {
    %x(docker #{state ? "start" : "stop"} #{container} >/dev/null 2>&1)
    container_state = %x(docker inspect -f {{.State.Running}} #{container} 2>/dev/null)
    expect($?).to be_success
    expect(container_state.strip).to eq(state ? "true" : "false")
  }
end
