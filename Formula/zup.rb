class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.6/zup.tar.gz"
  sha256 "699612b1e787a47ebb94d48a5789315925d12f4db718d9e1316a0c0865725538"
  version "0.1.6"

  def install
    bin.install "zup"
  end
end
