class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.2/zup.tar.gz"
  sha256 "b63f4256a77f79a7dfc322c6a95593c1364dee60834749a83d8e99517afdc976"
  version "0.1.4"

  def install
    bin.install "zup"
  end
end
