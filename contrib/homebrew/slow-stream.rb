require "language/go"

class SlowStream < Formula
  desc "slow-stream: pipe to throttle streams"
  homepage "https://github.com/moul/slow-stream"
  url "https://github.com/moul/slow-stream/archive/v1.0.0.tar.gz"
  sha256 "ff0313024168eaaadf5231889eb0145467db9ebf1a78d46ad678a73f7098d22c"

  head "https://github.com/moul/slow-stream.git"

  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["GOBIN"] = buildpath
    ENV["GO15VENDOREXPERIMENT"] = "1"
    (buildpath/"src/github.com/moul/slow-stream").install Dir["*"]

    system "go", "build", "-o", "#{bin}/slow-stream", "-v", "github.com/moul/slow-stream/cmd/slow-stream/"

    # FIXME: add autocompletion
  end

  test do
    output = shell_output(bin/"slow-stream --version")
    assert output.include? "slow-stream version "
  end
end
