function addContact() {

    let name = document.getElementById("inputName").value;
    let email = document.getElementById("inputEmail").value;
    let phone = document.getElementById("inputNumber").value;
    let subject = document.getElementById("inputSubject").value;
    let message = document.getElementById("inputMessage").value;


    if(name == "") {
        return alert ("Name should be filled");
    } else if(email == "") {
        return alert ("Email should be filled");
    } else if(phone == "") {
        return alert ("Phone should be filled");
    } else if(subject == "") {
        return alert ("Subject should be filled");
    } else if(message == "") {
        return alert ("Message should be filled");
    }

    
    let emailReceiver =  "dumbwaysIndonesia@gmail.com";

    let a = document.createElement("a")
    a.href = `mailto:${emailReceiver}?subject=${subject}&body=Hello, Saya ${name}, ${message}. Hubungi saya dinomer telephone ${phone}`;
    a.click()
}

