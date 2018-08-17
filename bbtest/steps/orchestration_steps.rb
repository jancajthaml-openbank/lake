require 'tempfile'

step "no :container :label is running" do |container, label|
  containers = %x(docker ps -a -f name=#{label} | awk '$2 ~ "#{container}" {print $1}' 2>/dev/null)
  expect($?).to be_success

  ids = containers.split("\n").map(&:strip).reject(&:empty?)
  return if ids.empty?

  ids.each { |id|
    eventually(timeout: 2) {
      send ":container running state is :state", id, false

      label = %x(docker inspect --format='{{.Name}}' #{id})
      label = ($? == 0 ? label.strip : id)

      %x(docker exec #{container} journalctl -o short-precise -u lake.service --no-pager >/reports/#{label}.log 2>&1)
      %x(docker rm -f #{id} &>/dev/null || :)
    }
  }
end

step ":container running state is :state" do |container, state|
  eventually(timeout: 3) {
    %x(docker exec #{container} systemctl stop lake.service 2>&1) unless state
    expect($?).to be_success

    %x(docker #{state ? "start" : "stop"} #{container} >/dev/null 2>&1)

    container_state = %x(docker inspect -f {{.State.Running}} #{container} 2>/dev/null)
    expect($?).to be_success
    expect(container_state.strip).to eq(state ? "true" : "false")
  }
end

step ":container :version is started with" do |container, version, label, params|
  containers = %x(docker ps -a --filter name=#{label} --filter status=running --format "{{.ID}} {{.Image}}")
  expect($?).to be_success
  containers = containers.split("\n").map(&:strip).reject(&:empty?)

  unless containers.empty?
    id, image = containers[0].split(" ")
    return if (image == "#{container}:#{version}")
  end

  send "no :container :label is running", container, label

  prefix = ENV.fetch('COMPOSE_PROJECT_NAME', "")
  my_id = %x(cat /etc/hostname).strip
  args = [
    "docker",
    "run",
    "-d",
    "-h #{label}",
    "--net=#{prefix}_default",
    "--volumes-from=#{my_id}",
    "--log-driver=json-file",
    "--net-alias=#{label}",
    "--name=#{label}",
    "--privileged"
  ] << params << [
    "#{container}:#{version}",
    "2>&1"
  ]

  id = %x(#{args.join(" ")})
  expect($?).to be_success, id

  eventually(timeout: 3) {
    send ":container running state is :state", id, true
  }
end

step "lake is running" do ||
  eventually(timeout: 10) {
    send ":container :version is started with", "openbankdev/lake_candidate", ENV.fetch("VERSION", "latest"), "lake", [
      "-v /sys/fs/cgroup:/sys/fs/cgroup:ro",
      "-p 5561",
      "-p 5562"
    ]
  }

  eventually(timeout: 5) {
    lake_handshake()
  }
end

step "lake is running with following configuration" do |configuration|
  eventually(timeout: 10) {
    send ":container :version is started with", "openbankdev/lake_candidate", ENV.fetch("VERSION", "latest"), "lake", [
      "-v /sys/fs/cgroup:/sys/fs/cgroup:ro",
      "-p 5561",
      "-p 5562"
    ]
  }

  params = configuration.split("\n").map(&:strip).reject(&:empty?).join("\n").inspect.delete('\"')

  containers = %x(docker ps -a --filter name=lake --filter status=running --format "{{.ID}}")
  expect($?).to be_success
  containers = containers.split("\n").map(&:strip).reject(&:empty?)

  expect(containers).not_to be_empty

  id = containers[0]

  %x(docker exec #{id} bash -c "echo -e '#{params}' > /etc/init/lake.conf" 2>&1)
  %x(docker exec #{id} systemctl restart lake.service 2>&1)

  eventually(timeout: 5) {
    lake_handshake()
  }
end
