
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/private';
 * 
 * console.log(ENVIRONMENT); // => "production"
 * console.log(PUBLIC_BASE_URL); // => throws error during build
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/private' {
	export const SHELL: string;
	export const npm_command: string;
	export const USER_ZDOTDIR: string;
	export const __ETC_PROFILE_DONE: string;
	export const npm_config_userconfig: string;
	export const COLORTERM: string;
	export const __HM_SESS_VARS_SOURCED: string;
	export const HYPRLAND_CMD: string;
	export const XDG_CONFIG_DIRS: string;
	export const npm_config_cache: string;
	export const XDG_SESSION_PATH: string;
	export const NIX_BUILD_CORES: string;
	export const TERM_PROGRAM_VERSION: string;
	export const configureFlags: string;
	export const XDG_BACKEND: string;
	export const mesonFlags: string;
	export const shell: string;
	export const depsHostHost: string;
	export const NODE: string;
	export const SSH_AUTH_SOCK: string;
	export const STRINGS: string;
	export const depsTargetTarget: string;
	export const XCURSOR_PATH: string;
	export const stdenv: string;
	export const COLOR: string;
	export const LOCALE_ARCHIVE_2_27: string;
	export const npm_config_local_prefix: string;
	export const builder: string;
	export const DESKTOP_SESSION: string;
	export const SSH_AGENT_PID: string;
	export const __ETC_PROFILE_NIX_SOURCED: string;
	export const HL_INITIAL_WORKSPACE_TOKEN: string;
	export const NO_AT_BRIDGE: string;
	export const npm_config_globalconfig: string;
	export const EDITOR: string;
	export const XDG_SEAT: string;
	export const PWD: string;
	export const NIX_PROFILES: string;
	export const SOURCE_DATE_EPOCH: string;
	export const LOGNAME: string;
	export const XDG_SESSION_DESKTOP: string;
	export const QT_QPA_PLATFORMTHEME: string;
	export const XDG_SESSION_TYPE: string;
	export const NIX_ENFORCE_NO_NATIVE: string;
	export const CUPS_DATADIR: string;
	export const MANPATH: string;
	export const NIX_PATH: string;
	export const npm_config_init_module: string;
	export const NIXPKGS_CONFIG: string;
	export const enableParallelInstalling: string;
	export const CXX: string;
	export const _: string;
	export const TEMPDIR: string;
	export const system: string;
	export const VSCODE_GIT_ASKPASS_NODE: string;
	export const HOST_PATH: string;
	export const QT_STYLE_OVERRIDE: string;
	export const VSCODE_INJECTION: string;
	export const enableParallelBuilding: string;
	export const GTK2_RC_FILES: string;
	export const IN_NIX_SHELL: string;
	export const doInstallCheck: string;
	export const buildCommandPath: string;
	export const HOME: string;
	export const enableParallelChecking: string;
	export const NIX_BINTOOLS: string;
	export const SSH_ASKPASS: string;
	export const LANG: string;
	export const NIXOS_OZONE_WL: string;
	export const _JAVA_AWT_WM_NONREPARENTING: string;
	export const LS_COLORS: string;
	export const XDG_CURRENT_DESKTOP: string;
	export const depsTargetTargetPropagated: string;
	export const KREW_ROOT: string;
	export const npm_package_version: string;
	export const STARSHIP_SHELL: string;
	export const WAYLAND_DISPLAY: string;
	export const cmakeFlags: string;
	export const STARSHIP_CONFIG: string;
	export const GIO_EXTRA_MODULES: string;
	export const outputs: string;
	export const NIX_STORE: string;
	export const TMPDIR: string;
	export const GIT_ASKPASS: string;
	export const XDG_SEAT_PATH: string;
	export const LD: string;
	export const INIT_CWD: string;
	export const READELF: string;
	export const CHROME_DESKTOP: string;
	export const STARSHIP_SESSION_KEY: string;
	export const NIX_USER_PROFILE_DIR: string;
	export const INFOPATH: string;
	export const npm_lifecycle_script: string;
	export const doCheck: string;
	export const VSCODE_GIT_ASKPASS_EXTRA_ARGS: string;
	export const npm_config_npm_version: string;
	export const depsBuildBuild: string;
	export const XDG_SESSION_CLASS: string;
	export const TERM: string;
	export const npm_package_name: string;
	export const GTK_PATH: string;
	export const SIZE: string;
	export const propagatedNativeBuildInputs: string;
	export const npm_config_prefix: string;
	export const ZDOTDIR: string;
	export const USER: string;
	export const strictDeps: string;
	export const VSCODE_GIT_IPC_HANDLE: string;
	export const TZDIR: string;
	export const AR: string;
	export const AS: string;
	export const PAM_KWALLET5_LOGIN: string;
	export const HYPRLAND_INSTANCE_SIGNATURE: string;
	export const TEMP: string;
	export const DISPLAY: string;
	export const NIX_BINTOOLS_WRAPPER_TARGET_HOST_x86_64_unknown_linux_gnu: string;
	export const npm_lifecycle_event: string;
	export const SHLVL: string;
	export const MOZ_ENABLE_WAYLAND: string;
	export const NIX_BUILD_TOP: string;
	export const NM: string;
	export const PAGER: string;
	export const __HM_ZSH_SESS_VARS_SOURCED: string;
	export const NIX_CFLAGS_COMPILE: string;
	export const QTWEBKIT_PLUGIN_PATH: string;
	export const patches: string;
	export const passAsFile: string;
	export const __NIXOS_SET_ENVIRONMENT_DONE: string;
	export const XDG_VTNR: string;
	export const buildInputs: string;
	export const XDG_SESSION_ID: string;
	export const LOCALE_ARCHIVE: string;
	export const LESSKEYIN_SYSTEM: string;
	export const npm_config_user_agent: string;
	export const QML2_IMPORT_PATH: string;
	export const TERMINFO_DIRS: string;
	export const npm_execpath: string;
	export const LD_LIBRARY_PATH: string;
	export const XDG_RUNTIME_DIR: string;
	export const NODE_PATH: string;
	export const depsBuildTarget: string;
	export const OBJCOPY: string;
	export const NIX_XDG_DESKTOP_PORTAL_DIR: string;
	export const out: string;
	export const npm_package_json: string;
	export const VSCODE_GIT_ASKPASS_MAIN: string;
	export const STRIP: string;
	export const XDG_DATA_DIRS: string;
	export const GDK_BACKEND: string;
	export const LIBEXEC_PATH: string;
	export const TMP: string;
	export const OBJDUMP: string;
	export const npm_config_noproxy: string;
	export const PATH: string;
	export const propagatedBuildInputs: string;
	export const npm_config_node_gyp: string;
	export const CC: string;
	export const NIX_CC: string;
	export const ORIGINAL_XDG_CURRENT_DESKTOP: string;
	export const DBUS_SESSION_BUS_ADDRESS: string;
	export const depsBuildTargetPropagated: string;
	export const FZF_DEFAULT_OPTS: string;
	export const depsBuildBuildPropagated: string;
	export const npm_config_global_prefix: string;
	export const NIX_CC_WRAPPER_TARGET_HOST_x86_64_unknown_linux_gnu: string;
	export const KPACKAGE_DEP_RESOLVERS_PATH: string;
	export const QT_PLUGIN_PATH: string;
	export const CONFIG_SHELL: string;
	export const __structuredAttrs: string;
	export const GIO_LAUNCHED_DESKTOP_FILE_PID: string;
	export const npm_node_execpath: string;
	export const RANLIB: string;
	export const GIO_LAUNCHED_DESKTOP_FILE: string;
	export const NIX_HARDENING_ENABLE: string;
	export const OLDPWD: string;
	export const NIX_LDFLAGS: string;
	export const nativeBuildInputs: string;
	export const name: string;
	export const TERM_PROGRAM: string;
	export const depsHostHostPropagated: string;
	export const NODE_ENV: string;
}

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/public';
 * 
 * console.log(ENVIRONMENT); // => throws error during build
 * console.log(PUBLIC_BASE_URL); // => "http://site.com"
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * 
 * console.log(env.ENVIRONMENT); // => "production"
 * console.log(env.PUBLIC_BASE_URL); // => undefined
 * ```
 */
declare module '$env/dynamic/private' {
	export const env: {
		SHELL: string;
		npm_command: string;
		USER_ZDOTDIR: string;
		__ETC_PROFILE_DONE: string;
		npm_config_userconfig: string;
		COLORTERM: string;
		__HM_SESS_VARS_SOURCED: string;
		HYPRLAND_CMD: string;
		XDG_CONFIG_DIRS: string;
		npm_config_cache: string;
		XDG_SESSION_PATH: string;
		NIX_BUILD_CORES: string;
		TERM_PROGRAM_VERSION: string;
		configureFlags: string;
		XDG_BACKEND: string;
		mesonFlags: string;
		shell: string;
		depsHostHost: string;
		NODE: string;
		SSH_AUTH_SOCK: string;
		STRINGS: string;
		depsTargetTarget: string;
		XCURSOR_PATH: string;
		stdenv: string;
		COLOR: string;
		LOCALE_ARCHIVE_2_27: string;
		npm_config_local_prefix: string;
		builder: string;
		DESKTOP_SESSION: string;
		SSH_AGENT_PID: string;
		__ETC_PROFILE_NIX_SOURCED: string;
		HL_INITIAL_WORKSPACE_TOKEN: string;
		NO_AT_BRIDGE: string;
		npm_config_globalconfig: string;
		EDITOR: string;
		XDG_SEAT: string;
		PWD: string;
		NIX_PROFILES: string;
		SOURCE_DATE_EPOCH: string;
		LOGNAME: string;
		XDG_SESSION_DESKTOP: string;
		QT_QPA_PLATFORMTHEME: string;
		XDG_SESSION_TYPE: string;
		NIX_ENFORCE_NO_NATIVE: string;
		CUPS_DATADIR: string;
		MANPATH: string;
		NIX_PATH: string;
		npm_config_init_module: string;
		NIXPKGS_CONFIG: string;
		enableParallelInstalling: string;
		CXX: string;
		_: string;
		TEMPDIR: string;
		system: string;
		VSCODE_GIT_ASKPASS_NODE: string;
		HOST_PATH: string;
		QT_STYLE_OVERRIDE: string;
		VSCODE_INJECTION: string;
		enableParallelBuilding: string;
		GTK2_RC_FILES: string;
		IN_NIX_SHELL: string;
		doInstallCheck: string;
		buildCommandPath: string;
		HOME: string;
		enableParallelChecking: string;
		NIX_BINTOOLS: string;
		SSH_ASKPASS: string;
		LANG: string;
		NIXOS_OZONE_WL: string;
		_JAVA_AWT_WM_NONREPARENTING: string;
		LS_COLORS: string;
		XDG_CURRENT_DESKTOP: string;
		depsTargetTargetPropagated: string;
		KREW_ROOT: string;
		npm_package_version: string;
		STARSHIP_SHELL: string;
		WAYLAND_DISPLAY: string;
		cmakeFlags: string;
		STARSHIP_CONFIG: string;
		GIO_EXTRA_MODULES: string;
		outputs: string;
		NIX_STORE: string;
		TMPDIR: string;
		GIT_ASKPASS: string;
		XDG_SEAT_PATH: string;
		LD: string;
		INIT_CWD: string;
		READELF: string;
		CHROME_DESKTOP: string;
		STARSHIP_SESSION_KEY: string;
		NIX_USER_PROFILE_DIR: string;
		INFOPATH: string;
		npm_lifecycle_script: string;
		doCheck: string;
		VSCODE_GIT_ASKPASS_EXTRA_ARGS: string;
		npm_config_npm_version: string;
		depsBuildBuild: string;
		XDG_SESSION_CLASS: string;
		TERM: string;
		npm_package_name: string;
		GTK_PATH: string;
		SIZE: string;
		propagatedNativeBuildInputs: string;
		npm_config_prefix: string;
		ZDOTDIR: string;
		USER: string;
		strictDeps: string;
		VSCODE_GIT_IPC_HANDLE: string;
		TZDIR: string;
		AR: string;
		AS: string;
		PAM_KWALLET5_LOGIN: string;
		HYPRLAND_INSTANCE_SIGNATURE: string;
		TEMP: string;
		DISPLAY: string;
		NIX_BINTOOLS_WRAPPER_TARGET_HOST_x86_64_unknown_linux_gnu: string;
		npm_lifecycle_event: string;
		SHLVL: string;
		MOZ_ENABLE_WAYLAND: string;
		NIX_BUILD_TOP: string;
		NM: string;
		PAGER: string;
		__HM_ZSH_SESS_VARS_SOURCED: string;
		NIX_CFLAGS_COMPILE: string;
		QTWEBKIT_PLUGIN_PATH: string;
		patches: string;
		passAsFile: string;
		__NIXOS_SET_ENVIRONMENT_DONE: string;
		XDG_VTNR: string;
		buildInputs: string;
		XDG_SESSION_ID: string;
		LOCALE_ARCHIVE: string;
		LESSKEYIN_SYSTEM: string;
		npm_config_user_agent: string;
		QML2_IMPORT_PATH: string;
		TERMINFO_DIRS: string;
		npm_execpath: string;
		LD_LIBRARY_PATH: string;
		XDG_RUNTIME_DIR: string;
		NODE_PATH: string;
		depsBuildTarget: string;
		OBJCOPY: string;
		NIX_XDG_DESKTOP_PORTAL_DIR: string;
		out: string;
		npm_package_json: string;
		VSCODE_GIT_ASKPASS_MAIN: string;
		STRIP: string;
		XDG_DATA_DIRS: string;
		GDK_BACKEND: string;
		LIBEXEC_PATH: string;
		TMP: string;
		OBJDUMP: string;
		npm_config_noproxy: string;
		PATH: string;
		propagatedBuildInputs: string;
		npm_config_node_gyp: string;
		CC: string;
		NIX_CC: string;
		ORIGINAL_XDG_CURRENT_DESKTOP: string;
		DBUS_SESSION_BUS_ADDRESS: string;
		depsBuildTargetPropagated: string;
		FZF_DEFAULT_OPTS: string;
		depsBuildBuildPropagated: string;
		npm_config_global_prefix: string;
		NIX_CC_WRAPPER_TARGET_HOST_x86_64_unknown_linux_gnu: string;
		KPACKAGE_DEP_RESOLVERS_PATH: string;
		QT_PLUGIN_PATH: string;
		CONFIG_SHELL: string;
		__structuredAttrs: string;
		GIO_LAUNCHED_DESKTOP_FILE_PID: string;
		npm_node_execpath: string;
		RANLIB: string;
		GIO_LAUNCHED_DESKTOP_FILE: string;
		NIX_HARDENING_ENABLE: string;
		OLDPWD: string;
		NIX_LDFLAGS: string;
		nativeBuildInputs: string;
		name: string;
		TERM_PROGRAM: string;
		depsHostHostPropagated: string;
		NODE_ENV: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://example.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.ENVIRONMENT); // => undefined, not public
 * console.log(env.PUBLIC_BASE_URL); // => "http://example.com"
 * ```
 * 
 * ```
 * 
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
