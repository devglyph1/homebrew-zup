class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.3/zup.tar.gz"
  sha256 "6b12fea08be0d8cd50d93a259e441e92c9a708607fb0d2ed0dfdde732bc9f78f"
  version "0.1.5"

  def install
    bin.install "zup"
  end
end
