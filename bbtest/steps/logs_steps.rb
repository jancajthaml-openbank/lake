
step "journalctl of :unit contains following" do |unit, expected|

  expected_lines = expected.split("\n").map(&:strip).reject(&:empty?)

  containers = %x(docker ps -a -f status=running -f name=lake | awk '{ print $1,$2 }' | sed 1,1d)
  expect($?).to be_success
  containers = containers.split("\n").map(&:strip).reject(&:empty?)

  expect(containers).not_to be_empty

  id = containers[0].split(" ")[0]

  # fixme be wary of duplicates in logs

  with_deadline(timeout: 5) {
    eventually(timeout: 2) {
      actual = %x(docker exec #{id} journalctl -o short-precise -u #{unit} --no-pager 2>&1)
      expect($?).to be_success

      actual_lines_merged = actual.split("\n").map(&:strip).reject(&:empty?)
      actual_lines = []
      idx = actual_lines_merged.length - 1

      loop do
        break if idx < 0 or actual_lines_merged[idx].include? "Started openbank lake message relay."
        actual_lines << actual_lines_merged[idx]
        idx -= 1
      end

      expected_lines.each { |line|
        found = false
        actual_lines.each { |l|
          next unless l.include? line
          found = true
          break
        }
        raise "#{line} was not found in logs:\n#{actual_lines.join("\n")}" unless found
      }
    }
  }
end
