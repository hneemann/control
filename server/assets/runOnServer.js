function slider(n) {
    let slValues=""
    for (let i = 0; i < n; i++) {
        let sl = document.getElementById('slider-' + i);
        if (slValues !== "") {
            slValues += ",";
        }
        slValues += sl.value;
    }

    let source = document.getElementById('source');
    let result = document.getElementById('slider-inner');

    let formData = new FormData();
    formData.append('data', source.value);
    formData.append('slider', slValues);

    fetchHelperForm("/execute/", formData, a => {
        result.innerHTML = a;
    });
}

function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');

    fetchHelper("/execute/", source.value, a => {
        result.innerHTML = a;
        source.focus();
    });
}

let myWindow;

function runSourceInWindow() {
    let source = document.getElementById('source');

    fetchHelper("/execute/", source.value, a => {
        if (myWindow && !myWindow.closed) {
            myWindow.document.body.innerHTML = "";
        } else {
            myWindow = window.open("", "", "width=810,height=620");
        }

        myWindow.document.write("<head>\n" +
            "  <meta charset=\"UTF-8\">\n" +
            "  <title>Control</title>\n" +
            "  <link rel=\"icon\" type=\"image/svg\" href=\"/assets/icon.svg\">\n" +
            "  <link rel=\"stylesheet\" type=\"text/css\" href=\"/assets/window.css\"/>\n" +
            "  <script type=\"text/javascript\" src=\"/assets/runOnServer.js\"></script>\n"+
            "  <script type=\"text/javascript\" src=\"/assets/main.js\"></script>\n"+
            "</head>\n" +
            "<body>\n"+
            "<textarea id=\"source\" style=\"display:none;\">" + source.value + "</textarea>\n");
        myWindow.document.write(a);
        myWindow.document.write("\n</body>\n</html>");
        source.focus();
    });
}
