const fs = require('fs');
function urlencodedForm(details) {
    var fd = new FormData();
    for (let property in details) {
        fd.append(property, details[property]);
    }
    var s = '';
    function encode(s){ return encodeURIComponent(s).replace(/%20/g,'+'); }
    for(var pair of fd.entries()){
        if(typeof pair[1]=='string'){
            s += (s?'&':'') + encode(pair[0])+'='+encode(pair[1]);
        }
    }
    return s;
}

function getLevel(id) {
    var formBody = urlencodedForm({
        'id': id
    });
    
    return fetch("http://delugedrop.com/3Dash/get_json.php", { method: "POST", body: formBody, headers: {"Content-Type": "application/x-www-form-urlencoded"} })
}

function getRecent() {
    return fetch("http://delugedrop.com/3Dash/get_recent.php");
}

function pushLevel(formdata) {
    // I am not gonna bother to require every argument in that formdata
    data = urlencodedForm(formdata);
    //this doesnt even work, not gonna even bother about making it work.... Since it isn't about this repo anyway

    console.log(data);
    return fetch("http://delugedrop.com/3Dash/push_level_data.php", { method: "POST", body: data, headers: {"Content-Type": "application/x-www-form-urlencoded"}});
}

//getLevel(5744).then(res => res.json().then(json => fs.writeFileSync('./a.json', JSON.stringify(json, 0, 4)))); //5692
//getRecent().then(res => res.text().then(text => console.log(text)));
pushLevel({"name": "If not the first IMPOSSIBLE level in 3Dash", "author": "Proudly, RewardedIvan", "difficulty": 5, "data": JSON.stringify(JSON.parse(fs.readFileSync("./lvl.json")))}).then(res => res.text().then(text => console.log(text)));
// Once again, some spaghetti code for fetching online levels and posting