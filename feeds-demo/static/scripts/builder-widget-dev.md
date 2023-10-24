
<!-- START: Builder Widget -->

<section id="builder-widget" class="widget">
<!-- This is where "the widget" should display -->
</section>

<noscript>JavaScript is required to display and use the Builder Widget</noscript>

<script src="CL-core.js"></script>
<script src="CL-feeds.js"></script>
<script src="CL-feeds-ui.js"></script>

<script src="CL-BuilderWidget.js"></script>

<script>
(function (document, window) {
    let cl = Object.assign({}, window.CL),
        widget_element = document.getElementById("builder-widget");

    /* NOTE: We want the builder to be hosted
     * where our code is deployed */
    cl.BaseURL = "";
    cl.BuilderWidget(widget_element);
}(document, window));
</script>



<!--   END: Builder Widget -->

- [CL-Builder-Widget.js](CL-Builder-Widget.js)
- [CL-core.js](CL-core.js)
- [CL-feeds.js](CL-feeds.js)
- [CL-feeds-ui.js](CL-feeds-ui.js)

