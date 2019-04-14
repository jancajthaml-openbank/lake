require_relative 'placeholders'

require 'json'

step "metrics file :path should have following keys:" do |path, keys|
  expected_keys = keys.split("\n").map(&:strip).reject { |x| x.empty? }

  eventually(timeout: 3) {
    expect(File.file?(path)).to be(true)
  }

  metrics_keys = File.open(path, 'rb') { |f| JSON.parse(f.read).keys }

  expect(metrics_keys).to match_array(expected_keys)
end
