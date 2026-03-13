import adapter from "@sveltejs/adapter-static";

const config = {
  kit: {
    adapter: adapter({
      pages: "../pb_public",
      assets: "../pb_public",
      fallback: "index.html",
    }),
  },
};

export default config;
