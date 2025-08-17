function updateByGui(n) {
    let slValues=""
    for (let i = 0; i < n; i++) {
        let sl = document.getElementById('guiElement-' + i);
        if (slValues !== "") {
            slValues += ",";
        }
        slValues += sl.value;
    }

    let source = document.getElementById('gui-source');
    let result = document.getElementById('gui-inner');

    let formData = new FormData();
    formData.append('data', source.value);
    formData.append('gui', slValues);

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
            "<div class=\"div-main\">\n");
        myWindow.document.write(a);
        myWindow.document.write("\n</div></body>\n</html>");
        source.focus();
    });
}
