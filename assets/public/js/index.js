function setBundles() {
  const plugins = [];
  $("#pluginsList")
    .children()
    .each((index, element) => {
      var current = $(element);
      var txt = current.text();
      if (txt != "") {
        plugins.push(txt);
      }
    });
  $("#pluginsHidden").val(plugins.join(","));
}

$(document).ready(() => {
  $(window).on("scroll load", function () {
    if ($(window).scrollTop() > 1000) {
      $("#scrollTop").addClass("active");
    } else {
      $("#scrollTop").removeClass("active");
    }
  });

  $("#scrollTop").on("click", function (e) {
    e.preventDefault();
    $("html, body").animate({ scrollTop: 0 }, 1000);
  });

  $(document).on("click", "#pluginRemove", (e) => {
    e.preventDefault();
    $(e.target).parents("#pluginListItem").remove();
    setBundles();
  });

  $("#pluginAdd").click((e) => {
    e.preventDefault();
    const val = $("#pluginInput").val();
    $("#pluginsList").append(
      `<li id="pluginListItem" class="border border-secondary rounded-pill px-3 py-2"><span>` +
        val +
        `<button id="pluginRemove"><i class="bi bi-dash-circle ml-3"></i></button></span></li>`
    );
    $("#pluginInput").val("");
    setBundles();
  });
});
