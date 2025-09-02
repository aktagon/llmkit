class Llmkit < Formula
  desc "Command-line toolkit for working with Large Language Models"
  homepage "https://github.com/aktagon/llmkit"
  # NOTE: The url, version, and sha256 are updated by the github action (.github/workflows/release.yml) automatically
  url "https://github.com/aktagon/llmkit/archive/refs/tags/v0.2.6.tar.gz"
  version "v0.2.6"
  sha256 "644266e1b63537d557b342f7f9db26daa966261003206df626cbed7f18eea41f"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"llmkit", "./cmd/llmkit"
  end

  def caveats
    <<~EOS
      llmkit provides unified access to multiple LLM providers.

      Set your API keys as environment variables:
        export OPENAI_API_KEY="your-openai-key"
        export ANTHROPIC_API_KEY="your-anthropic-key"
        export GOOGLE_API_KEY="your-google-key"

      You can add these to your shell profile (~/.zshrc, ~/.bashrc, etc.)
      to make them permanent.

      Use 'llmkit --help' to see available commands and options.
    EOS
  end

  test do
    # Test that the binary was installed and can display help
    assert_match "llmkit", shell_output("#{bin}/llmkit --help 2>&1", 1)
  end
end