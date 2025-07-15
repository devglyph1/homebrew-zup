class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/zup"
  url "https://github.com/devglyph1/zup/releases/download/v0.1.2/zup.tar.gz"
  sha256 "5fe8120ebcb2eed6e234afeaf33d1f3b301f8da74d31b4f89dbd798739a0d586"
  version "0.1.1"

  def install
    bin.install "zup"
  end
end
