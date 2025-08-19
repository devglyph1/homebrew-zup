class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.12/zup.tar.gz"
  sha256 "536909a67e29f375293ef0ab1ab5db1884441d7f5665c7b547f24c9223b9ce3f"
  version "0.1.12"

  def install
    bin.install "zup"
  end
end
