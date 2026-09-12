document$.subscribe(function () {
  if (typeof mermaid === "undefined") {
    return;
  }

  mermaid.initialize({
    startOnLoad: false,
    securityLevel: "strict",
    theme: document.body.getAttribute("data-md-color-scheme") === "slate" ? "dark" : "base"
  });

  mermaid.run({
    nodes: document.querySelectorAll(".mermaid")
  });
});
