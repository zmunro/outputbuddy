
class Outputbuddy < Formula
  desc "Flexible output redirection with color preservation"
  homepage "https://github.com/zmunro/outputbuddy"
  version "2.0.0"

  # TODO: Update the URL after creating your first release
  url "https://github.com/zmunro/outputbuddy/archive/refs/tags/v#{version}.tar.gz"
  sha256 "PLACEHOLDER_SHA256"  # This will be updated when you create the release

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
