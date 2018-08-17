
step "systemctl contains following" do |packages|
  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  items = packages.split("\n").map(&:strip).reject(&:empty?)

  ids.each { |id|
    eventually() {
      items.each { |item|
        units = %x(docker exec #{id} systemctl list-units --type=service | grep #{item} | awk '{ print $1 }')
        units = units.split("\n").map(&:strip).reject(&:empty?)
        expect(units).not_to be_empty, "#{item} was not found in #{id}"
      }
    }
  }
end

step "systemctl does not contains following" do |packages|
  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  items = packages.split("\n").map(&:strip).reject(&:empty?)

  ids.each { |id|
    items.each { |item|
      units = %x(docker exec #{id} systemctl list-units --type=service | grep #{item} | awk '{ print $1 }')
      units = units.split("\n").map(&:strip).reject(&:empty?)
      expect(units).to be_empty, "#{item} was not found in #{id}"
    }
  }
end

step ":operation unit :unit" do |operation, unit|
  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually(timeout: 5) {
      %x(docker exec #{id} systemctl #{operation} #{unit} 2>&1)
    }

    unless $? == 0
      err = %x(docker exec #{id} systemctl status #{unit} 2>&1)
      raise "operation \"systemctl #{operation} #{unit}\" returned error: #{err}"
    end
  }
end

step "unit :unit is running" do |unit|
  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually() {
      expect(Docker.unit_running?(id, unit)).to eq(true)
    }
  }
end

step "unit :unit is not running" do |unit|
  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  ids.each { |id|
    eventually() {
      expect(Docker.unit_running?(id, unit)).not_to eq(true)
    }
  }
end
