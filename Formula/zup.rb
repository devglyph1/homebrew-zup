class Zup < Formula
  desc "A fast, customizable CLI built with Cobra to automate local development environment setup — from cloning repos to installing tools and running services."
  homepage "https://github.com/devglyph1/homebrew-zup"
  url "https://github.com/devglyph1/homebrew-zup/releases/download/v0.1.1/zup.tar.gz"
  sha256 "22fbae5b510a538d2066f755cb8207914d3d95a7e426c0e304fedc2cd2383ce1"
  version "0.1.2"

  def install
    bin.install "zup"
  end
end
