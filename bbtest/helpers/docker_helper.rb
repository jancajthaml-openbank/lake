module Docker

  def self.get_journal(id, unit)
    data = %x(docker exec #{id} journalctl -o short-precise -u #{unit}.service --no-pager 2>&1)
    return data, $? == 0
  end

  def self.save_journal(id, unit, path)
    %x(docker exec #{id} journalctl -o short-precise -u #{unit}.service --no-pager >#{path} 2>&1)
    return $? == 0
  end

  def self.running?(id)
    out = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
    return false unless $? == 0
    return out.strip == "true"
  end

  def self.get_lakes()
    version = ENV.fetch("VERSION", "latest")
    containers = %x(docker ps -a --filter name=lake --filter status=running --format "{{.ID}} {{.Image}}")
    return [] unless $? == 0
    return containers.split("\n").map(&:strip).reject { |x| x.empty? or x.split(" ").last != "openbankdev/lake_candidate:#{version}" }.map { |x| x.split(" ").first }
  end

  def self.unit_enabled?(id, unit)
    %x(docker exec #{id} systemctl is-enabled #{unit} 2>&1)
    return $? == 0
  end

  def self.unit_running?(id, unit)
    out = %x(docker exec #{id} systemctl show -p SubState #{unit} 2>&1 | sed 's/SubState=//g')
    return out.strip == "running"
  end

  def unit_enabled?(id, unit)
    Docker.unit_enabled?(id, unit)
  end

  def unit_running?(id, unit)
    Docker.unit_running?(id, unit)
  end

  def unit_enable(id, unit)
    Docker.unit_enable(id, unit)
  end

  def get_journal(id, unit)
    Docker.get_journal(id, unit)
  end

  def save_journal(id, unit, path)
    Docker.save_journal(id, unit, path)
  end

  def get_lakes()
    Docker.get_lakes()
  end

  def running?(id)
    Docker.running?(id)
  end

end
