<!DOCTYPE html>
<html lang="ja-JP">

<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>Login</title>

    <link href="https://storage.googleapis.com/id.nerocrux.com/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://storage.googleapis.com/id.nerocrux.com/css/style.css" rel="stylesheet">

</head>

<body class="gray-bg">
    <div class="middle-box text-center loginscreen animated fadeInDown">

        {{template "content" . }}

        <script>
            $(document).ready(() => {
                $('#mainContainer').hide();
                checkIfLoggedIn()
                    .then((response) => {
                        if (response)
                            return loadMainContainer()
                    })
            })
        </script>
    </div>
</body>

</html>
