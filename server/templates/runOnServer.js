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
            "  <link rel=\"stylesheet\" type=\"text/css\" href=\"/assets/main.css\"/>\n" +
            "</head>\n" +
            "<body>\n");
        myWindow.document.write(a);
        myWindow.document.write("\n</body>\n</html>");
        source.focus();
    });
}
