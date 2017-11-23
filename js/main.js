$(function() {
  $('body').tooltip({
    selector: "[data-toggle='tooltip']"
  });

  $('body').popover({
    selector: "[data-toggle='popover']"
  });

  AOS.init({
    offset: 100,
    once: true
  });

  // Topbar dropdown custom scrollbar
  $(".navbar .menu").slimscroll({
    height: "200px",
    alwaysVisible: false,
    size: "3px"
  }).css("width", "100%")

  // Theme Customization
  $(".color-scheme-picker > li").on("click", function(e) {
    e.preventDefault();
    var $this = $(this);
    var availableColors = "skin-deep-blue skin-pale-green skin-deep-purple skin-deep-orange skin-blue-grey";

    $("body")
      .removeClass(availableColors)
      .addClass($this.data("color"));

    $("#customizationModal").modal("hide");
  })

  $(".sidebar-image-picker > li").on("click", function(e) {
    e.preventDefault();
    var $this = $(this);
    var availableImages = "rocky-beach dark-waves starry misty-mountains sunrise starry-mountains";

    $("body")
      .removeClass(availableImages)
      .addClass($this.data("image"));

    $("#customizationModal").modal("hide");
  })


  // Sidebar custom scrollbar
  if (!$("body").hasClass("fixed")) {
    if (typeof $.fn.slimScroll != 'undefined') {
      $(".sidebar").slimScroll({destroy: true}).height("auto");
    }
  } else {
    $(window).on("resize", function() {
      $(".sidebar").slimScroll({destroy: true}).height("auto");
      $(".sidebar").slimscroll({
        height: ($(window).height() - $(".main-header").height()) + "px",
        color: "rgba(255, 255, 255,0.4)",
        size: "3px"
      });
    }).trigger("resize")
  }
})
