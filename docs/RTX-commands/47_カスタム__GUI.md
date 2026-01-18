# 第47章: カスタム  GUI

> 元PDFページ: 621-622

---

第 47 章
カスタム  GUI
カスタム GUIとは、ルーターの設定を行うための GUI (WWW ブラウザに対応するユーザインタフェース ) をユーザ
が独自に設計し組み込むことができる機能です。ルーターにはホストから HTTPで設定を転送するためのインタフ
ェースが用意されており、ユーザは JavaScript を使用して GUIを作成します。
ヤマハルーターには WWWブラウザ設定支援機能が搭載されていますが、ユーザごとに設定画面を変更することは
できませんでした。本機能では、カスタム GUIを複数組み込み、ログインするユーザによって画面を切り替えるこ
とが可能です。
47.1 カスタム  GUI を使用するか否かの設定
[書式 ]
httpd  custom-gui  use use
no httpd  custom-gui  use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
[説明 ]
カスタム  GUI を使用するか否かを設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
47.2 カスタム  GUI を使用するユーザの設定
[書式 ]
httpd  custom-gui  user [user] directory= path [index= name ]
no httpd  custom-gui  user [user...]
[設定値及び初期値 ]
•user
• [設定値 ] : ユーザ名
• [初期値 ] : -
•path
• [設定値 ] : 基点となるディレクトリの絶対パスまたは相対パス
• [初期値 ] : -
•name
• [設定値 ] : スラッシュ  '/' 止めの  URL でアクセスした場合に出力するファイル名
• [初期値 ] : index.html
[説明 ]
カスタム  GUI を使用するユーザを設定する。 http://( ルーターの  IP アドレス  )/にアクセスし、本コマンドで登録され
ているユーザ名でログインすると  http://( ルーターの  IP アドレス  )/custom/ user/にリダイレクトされる。
user を省略した場合には無名ユーザに対する設定となる。この場合の  URL は http://( ルーターの  IP アドレス  )/
custom/anonymous.user/ となる。
path には基点となるディレクトリを絶対パス、もしくは相対パスで指定する。相対パスで指定した場合、環境変数
PWD を基点としたパスと解釈される。 PWD は set コマンドで変更可能であり、初期値は  "/" である。
name  にはブラウザから  '/' 止めの  URL でアクセスした場合に表示するファイル名を指定する。コマンドリファレンス  | カスタム  GUI | 621

[ノート ]
本コマンドを設定する場合、 無名ユーザ以外は事前に  login user  コマンドでユーザを登録しておく必要がある。登録
されていないユーザに対して本コマンドを設定するとエラーになる。
RTX1300 、RTX840 では無名ユーザは登録できないため、 user は省略することはできない。
本コマンドが設定されているユーザは、ルーターに内蔵されている通常の  GUI にアクセスすることができない。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
47.3 カスタム  GUI の API を使用するか否かの設定
[書式 ]
httpd  custom-gui  api use use
no httpd  custom-gui  api use [use]
[設定値及び初期値 ]
•use
• [設定値 ] :
設定値 説明
on 使用する
off 使用しない
• [初期値 ] : off
[説明 ]
API 用の  URL "http://( ルーターの  IP アドレス  )/custom/api" に対する  POST リクエストを受け付けるか否かを設定す
る。
[ノート ]
API 用の  URL を使用するには、本コマンドに加えて httpd custom-gui use on  が設定されている必要がある。
本コマンドを  on にしても httpd custom-gui api password  コマンドを設定しなければ  API 用の  URL を使用すること
はできない。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830
47.4 カスタム  GUI の API にアクセスするためのパスワードの設定
[書式 ]
httpd  custom-gui  api password  password
no httpd  custom-gui  api password  [password ]
[設定値及び初期値 ]
•password
• [設定値 ] : パスワード
• [初期値 ] : -
[説明 ]
API 用の  URL へ POST リクエストを送信する際のパスワードを設定する。 32 文字以内で半角英数字を使用するこ
とができる。
例えば、 本コマンドでパスワードとして  doremi を設定した場合、 URL は http://( ルーターの  IP アドレス  )/custom/api?
password=doremi となる。
[適用モデル ]
RTX1300, RTX1220, RTX1210, RTX840, RTX830622 | コマンドリファレンス  | カスタム  GUI

