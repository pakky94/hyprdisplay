{
  inputs.flake-utils.url = "github:numtide/flake-utils";
  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        pname = "hyprdisplay";
        version = "0.0.1";
        src = ./hyprdisplay/.;
        frontend = pkgs.stdenv.mkDerivation (finalAttrs: {
          inherit pname version src;

          nativeBuildInputs = with pkgs; [
            nodejs
            pnpm.configHook
          ];

          pnpmDeps = pkgs.pnpm.fetchDeps {
            inherit (finalAttrs) pname version src;
            sourceRoot = "hyprdisplay/frontend";
            hash = "sha256-Qm/SS0LD3uC8Jjveyu61FZ1jVRpoyNsIu9nAwL5EzTI=";
          };

          sourceRoot = "hyprdisplay/frontend";

          buildPhase = ''
            runHook preBuild
            pnpm run build
            runHook postBuild
          '';

          installPhase = ''
            runHook preInstall
            mkdir $out/
            cp -r ./dist/* $out/
            runHook postInstall
          '';

          # meta = {
          #   description = "GUI program developed by vue3";
          #   license = with pkgs.lib.licenses; [ gpl3Plus ];
          #   maintainers = with pkgs.lib.maintainers; [ aucub ];
          #   platforms = pkgs.lib.platforms.linux;
          # };
        });
      in
      {
        packages.hyprdisplay = pkgs.buildGoModule {
          inherit pname version src;
        
          vendorHash = "sha256-PlGHZ6CaZknVTFEDqjotgh07Q/W9dH0tRU7FGh1utS8=";
        
          nativeBuildInputs = with pkgs; [
            wails
            pkg-config
            wrapGAppsHook3
            autoPatchelfHook
            copyDesktopItems
          ];
        
          buildInputs = with pkgs; [
            webkitgtk_4_0
            libsoup_3
          ];
        
          # desktopItems = [
          #   (makeDesktopItem {
          #     name = "GUI.for.Clash";
          #     exec = "GUI.for.Clash";
          #     icon = "GUI.for.Clash";
          #     genericName = "GUI.for.Clash";
          #     desktopName = "GUI.for.Clash";
          #     categories = [
          #       "Network"
          #     ];
          #     keywords = [
          #       "Proxy"
          #     ];
          #   })
          # ];
        
          postUnpack = ''
            cp -r ${frontend} $sourceRoot/frontend/dist
          '';
        
          # postPatch = ''
          #   sed -i '/exePath, err := os.Executable()/,+3d' bridge/bridge.go
          #   substituteInPlace bridge/bridge.go \
          #     --replace-fail "Env.BasePath = filepath.Dir(exePath)" "" \
          #     --replace-fail "Env.AppName = filepath.Base(exePath)" "Env.AppName = \"GUI.for.Clash\"
          #       Env.BasePath = filepath.Join(os.Getenv(\"HOME\"), \".config\", Env.AppName)" \
          #     --replace-fail 'exePath := Env.BasePath' 'exePath := "${placeholder "out"}/bin"'
          # '';
        
          buildPhase = ''
            runHook preBuild
            wails build -m -s -trimpath -skipbindings -devtools -tags webkit2_40 -o hyprdisplay
            runHook postBuild
          '';
        
          installPhase = ''
            runHook preInstall
            mkdir -p $out
            cp -r ./build/bin $out
            #cp build/appicon.png $out/share/pixmaps/GUI.for.Clash.png
            runHook postInstall
          '';
        
          # meta = {
          #   description = "Clash GUI program developed by vue3 + wails";
          #   homepage = "https://github.com/GUI-for-Cores/GUI.for.Clash";
          #   mainProgram = "GUI.for.Clash";
          #   license = with lib.licenses; [ gpl3Plus ];
          #   maintainers = with lib.maintainers; [ aucub ];
          #   platforms = lib.platforms.linux;
          # };
        };

        devShells.default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go
            gtk3
            nodejs_20
            nsis
            pnpm.configHook
            upx
            wails
            webkitgtk
          ];
        };
      }
    );
}
