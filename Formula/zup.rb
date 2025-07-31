class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.1/zup.tar.gz"
  sha256 "6a4d01254bf63ddf5e19e08abffcc5251ecf80953e3aa0087c1747678313fefa"
  version "0.1.3"

  def install
    bin.install "zup"
  end
end
