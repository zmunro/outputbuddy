
class Outputbuddy < Formula
  desc "Flexible output redirection with color preservation"
  homepage "https://github.com/zmunro/outputbuddy"
  version "2.1.0"

  url "https://github.com/zmunro/outputbuddy/archive/refs/tags/v#{version}.tar.gz"
  sha256 "5caa4c051bd1b2eb43b45a05cf42136d6dfc6aa73f6f5f79629c467d9e8d1dfc"

  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.VERSION=#{version}")

    # Create the 'ob' symlink
    bin.install_symlink "outputbuddy" => "ob"
  end

  test do
    # Test basic functionality
    output = shell_output("#{bin}/outputbuddy --version")
    assert_match version.to_s, output

    # Test command execution
    shell_output("#{bin}/outputbuddy 2+1=test.log -- echo 'test'")
    assert_predicate testpath/"test.log", :exist?
  end
end
