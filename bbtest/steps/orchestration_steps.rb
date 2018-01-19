
step "container :container_name should be running" do |container_name|
  prefix = ENV.fetch('COMPOSE_PROJECT_NAME', "")
  container_id = %x(docker ps -aqf "name=#{prefix}_#{container_name}" 2>/dev/null)
  expect($?).to eq(0), "error running `docker ps -aq`: err:\n #{container_name}"

  eventually(timeout: 5) {
    container_id.split("\n").each { |id|
      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to eq(0), "error running `docker inspect -f {{.State.Running}}`: err:\n #{id}"

      expect(container_state.strip).to eq("true")
    }
  }
end




# mapping of container port to host port
#name=server
#id=$(docker ps -aqf "name=${name}" 2>/dev/null)
#docker inspect -f '{{(index (index .NetworkSettings.Ports "8080/tcp") 0).HostPort}}' ${id}
