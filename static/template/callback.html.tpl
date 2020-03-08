{{define "content"}}
<div id="mainContainer">
    {{if eq .Err ""}}
    <div class="widget-head-color-box p-lg text-center" id="userinfo">
        <div class="m-b-md">
            <h4 class="font-bold no-margins">
                <ul class="list-group">
                    <li class="list-group-item">
                        <span class="badge badge-primary">Subject</span>
                        <p>{{.Subject}}</p>
                    </li>
                    <li class="list-group-item">
                        <span class="badge badge-primary">KeyID</span>
                        <p>{{.KeyID}}</p>
                    </li>
                    <li class="list-group-item">
                        <span class="badge badge-primary">Issuer</span>
                        <p>{{.Issuer}}</p>
                    </li>
                    {{range $val := .Audience}}
                    <li class="list-group-item ">
                        <span class="badge badge-info">Audience</span>
                        <p>{{$val}}</p>
                    </li>
                    {{end}}
                </ul>
            </h4>
        </div>
    </div>
    {{else}}
        <div class="widget-head-color-box p-lg text-center" id="error">
            <div class="m-b-md">
                <p class="font-bold no-margins">
                    {{.Err}}
                </p>
            </div>
        </div>
        {{end}}
        <a href="/"><button type="submit" class="btn btn-primary block full-width m-b" value="Login">Retry</button></a>
</div>
{{end}}