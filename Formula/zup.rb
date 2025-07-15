class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/zup"
  url "https://github.com/devglyph1/zup/releases/download/v0.1.0/zup.tar.gz"
  sha256 "31f4139ed97aa8739b6c89f954d81ecaa2c9e35c0caf98970bad8311d867e64c"
  version "0.1.0"

  def install
    bin.install "zup"
  end
end
