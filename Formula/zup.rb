class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.9/zup.tar.gz"
  sha256 "47f2bbae83895ea1e70bb409f6020bb453ff9539c76497ff689a1aa153267361"
  version "0.1.9"

  def install
    bin.install "zup"
  end
end
