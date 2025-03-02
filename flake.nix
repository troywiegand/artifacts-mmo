{
  description = "artifacts dev flake";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      devShell = pkgs.mkShell {
        name = "artifacts";

        buildInputs = with pkgs; [
          go
        ];
      };
    }
  );
}
