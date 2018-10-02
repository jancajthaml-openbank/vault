require_relative 'placeholders'

require 'json'

step "metrics for tenant :tenant should report :count created accounts" do |tenant, count|
  metrics_file = "/opt/vault/metrics/metrics.#{tenant}.json"

  eventually(timeout: 3) {
    expect(File.file?(metrics_file)).to be(true)
    metrics = File.open(metrics_file, 'rb') { |f| JSON.parse(f.read) }
    expect(metrics["createdAccounts"]).to eq(count)
  }
end
