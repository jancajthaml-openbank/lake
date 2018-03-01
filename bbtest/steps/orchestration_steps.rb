step "container :container_name should be started from scratch" do |container_name|
  prefix = ENV.fetch('COMPOSE_PROJECT_NAME', "")
  container_id = %x(docker ps -aqf "name=#{prefix}_#{container_name}" 2>/dev/null)
  expect($?).to be_success

  containers = container_id.split("\n").map(&:strip).reject(&:empty?)
  expect(containers).not_to be_empty

  containers.each { |id|
    %x(docker kill --signal="TERM" #{id} >/dev/null 2>&1)
    expect($?).to be_success
  }

  eventually(timeout: 3) {
    containers.each { |id|
      %x(docker start #{id} >/dev/null 2>&1)
      expect($?).to be_success

      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("true")
    }
  }
end

step "container :container_name should be running" do |container_name|
  prefix = ENV.fetch('COMPOSE_PROJECT_NAME', "")
  container_id = %x(docker ps -aqf "name=#{prefix}_#{container_name}" 2>/dev/null)
  expect($?).to be_success

  containers = container_id.split("\n").map(&:strip).reject(&:empty?)
  expect(containers).not_to be_empty

  eventually(timeout: 3) {
    containers.each { |id|
      %x(docker start #{id} >/dev/null 2>&1)
      expect($?).to be_success

      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("true")
    }
  }
end

