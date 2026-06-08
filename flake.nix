{
  description = "TUI news app written in Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    # Small helper for multi-system flakes.
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
      };

      pname = "foxden";
      version = "0.1.0";
    in {
      packages.default = pkgs.buildGoModule {
        inherit pname version;

        src = ./.;

        # First build will intentionally fail and tell you the real hash.
        # Replace this with the hash from the error message.
        vendorHash = "sha256-CowlwrEFU8x1f2d99uccRKSwjA3NhqNWL4yin1j0vNs=";

        # If your main package is under ./cmd/foxden
        subPackages = ["cmd/foxden"];

        # Useful if you use SQLite with CGO disabled via modernc.org/sqlite.
        env.CGO_ENABLED = "0";

        ldflags = [
          "-s"
          "-w"
          "-X main.version=${version}"
        ];

        meta = {
          description = "Terminal news reader and scraper frontend";
          mainProgram = pname;
        };
      };

      apps.default = flake-utils.lib.mkApp {
        drv = self.packages.${system}.default;
      };

      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          # Go toolchain
          go
          gopls
          gotools
          go-tools
          delve

          # Formatting/linting
          golangci-lint

          # Project tooling
          git
          just
          sqlite
          jq
          yq

          # Helpful for terminal UI debugging
          vhs

          # Optional HTTP/debugging tools
          curl
          httpie
        ];

        shellHook = ''
          echo "foxden Go dev shell"
          echo
          echo "Useful commands:"
          echo "  go mod init github.com/patjrobinson/foxden"
          echo "  go get github.com/charmbracelet/bubbletea"
          echo "  go get github.com/charmbracelet/lipgloss"
          echo "  go get github.com/charmbracelet/bubbles"
          echo "  go get github.com/mmcdole/gofeed"
          echo "  go get github.com/PuerkitoBio/goquery"
          echo "  go get modernc.org/sqlite"
          echo
          echo "  go run ./cmd/foxden"
          echo "  go test ./..."
          echo "  nix build"
        '';
      };

      formatter = pkgs.nixpkgs-fmt;
    });
}
