
step "journalctl of :unit contains following" do |unit, expected|

  expected_lines = expected.split("\n").map(&:strip).reject(&:empty?)

  containers = %x(docker ps -a --filter name=vault --filter status=running --format "{{.ID}}")
  expect($?).to be_success

  ids = containers.split("\n").map(&:strip).reject(&:empty?)
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually() {
      out, ok = Docker.get_journal(id, unit)
      expect(ok).to be true

      actual_lines_merged = out.split("\n").map(&:strip).reject(&:empty?)
      actual_lines = []
      idx = actual_lines_merged.length - 1

      loop do
        break if idx < 0 or actual_lines_merged[idx].include? ": Started"
        actual_lines << actual_lines_merged[idx]
        idx -= 1
      end

      actual_lines = actual_lines.reverse

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
