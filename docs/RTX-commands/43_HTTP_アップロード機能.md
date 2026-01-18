# 第43章: HTTP アップロード機能

> 元PDFページ: 584-587

---

第 43 章
HTTP アップロード機能
ヤマハルーター内の情報  ( 設定ファイルあるいは  SYSLOG) を指定した  HTTP サーバーにアップロードすることが
できる機能です。
複数拠点の設定ファイルやログの集中管理に使用することができます。
設定ファイルは show config  コマンドまたは show config N  コマンド、 SYSLOG はshow log  コマンドの実行結果がフ
ァイルとして保存されます。
保存したファイルの先頭には、実行したコマンド名が表示されます。
HTTP サーバーに複数のヤマハルーターからの情報を集める場合など、 ファイルをディレクトリ指定して格納するこ
とができます。
ディレクトリを指定する場合には http upload  コマンドで設定します。
この機能を使用するためには、 HTTP サーバー側での対応も必要です。
HTTP サーバーの  OS の種類には依存しません  (Windows 、UNIX、etc.) が、 UNIX 上の  HTTP サーバーを使用する場
合、 CGI スクリプトは  nobody ユーザー権限として実行されるため、生成されるファイルも  nobody ユーザー権限と
なります。 CGI 実行ディレクトリのパーミッションは、 [------rw-] を満たしておく必要があります。
HTTP サーバー側で動作させる必要のあるスクリプトファイル、および本機能に関する技術情報は以下に示す  URL
で公開されています。
https://www.rtpro.yamaha.co.jp
43.1 HTTP アップロードするファイルの設定
[書式 ]
http upload  type [config_no ] [directory /] filename
no http upload  type [...]
[設定値及び初期値 ]
•type
• [設定値 ] : 'config' or 'log'
• [初期値 ] : -
•config_no
• [設定値 ] : 0-4.2
• [初期値 ] : -
•directory
• [設定値 ] : 出力先のディレクトリ名
• [初期値 ] : -
•filename
• [設定値 ] : 出力先のファイル名
• [初期値 ] : -
[説明 ]
HTTP サーバーにアップロードする情報と、保存先のディレクトリ名及びファイル名を設定する。
指定したディレクトリ名でディレクトリを生成し、そのディレクトリ内に指定したファイル名のファイルを生成す
る ( 例： dir1/dir2/config.txt) 。
ディレクトリ名、ファイル名に以下を指定することもできる。
文字列 意味
%y 年 ( yyyy )584 | コマンドリファレンス  | HTTP アップロード機能

文字列 意味
%m 月 ( mm )
%d 日 ( dd )
%H 時 ( hh )
%M 分 ( mm )
%S 秒 ( ss )
%n シリアル番号
%a LAN1 の MAC アドレス
%P 機種名
ディレクトリ、ファイル名に  '%' を含む文字列を指定する場合は、 '%' を続けて指定する必要がある。
type に 'config' を指定したときのみ  config_no  が有効になり、 config_no  を省略した場合は起動中の  config がアップロ
ードの対象になる。
なお、 config_no  は設定ファイル多重機能に対応した機種でのみ設定することができる。
[設定例 ]
「(機種名 )/(シリアル番号 )/(LAN1のMACアドレス )/20100101/120000.txt 」というディレクトリとファイルをアップロ
ードする
# http upload config %P/%n/%a/%y%m%d/%H%M%S.txt
「%config.txt 」というファイルをアップロードする
# http upload config %%config.txt
[適用モデル ]
RTX1300, RTX1220, RTX1210
43.2 HTTP アップロード先  URL の設定
[書式 ]
http upload  url url
no http upload  url [url]
[設定値及び初期値 ]
•url
• [設定値 ] : アップロード先の  URL
• [初期値 ] : -
[説明 ]
HTTP アップロードで使用する  HTTP サーバーの  URL を設定する。
HTTP サーバーでは  cgi を許可するよう設定にする必要があり、アップロードを受け入れるための  cgi を実行させる
必要がある。
[適用モデル ]
RTX1300, RTX1220, RTX1210
43.3 HTTP アップロードを許可するか否かの設定
[書式 ]
http upload  permit  switch
no http upload  permit  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :コマンドリファレンス  | HTTP アップロード機能  | 585

設定値 説明
on HTTP アップロードを許可する
off HTTP アップロードを許可しない
• [初期値 ] : off
[説明 ]
HTTP アップロードを許可するか否かを設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210
43.4 HTTP アップロードのタイムアウト時間の設定
[書式 ]
http upload  timeout  time
no http upload  timeout  [time]
[設定値及び初期値 ]
•time
• [設定値 ] : 1-180[ 秒]
• [初期値 ] : 30
[説明 ]
HTTP アップロードでタイムアウトするまでの時間を設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210
43.5 HTTP アップロードのリトライの間隔と回数の設定
[書式 ]
http upload  retry  interval  interval  count
no http upload  retry  interval  [..]
[設定値及び初期値 ]
•interval
• [設定値 ] : 1-60[秒]
• [初期値 ] : 30
•count
• [設定値 ] : 1-10
• [初期値 ] : 5
[説明 ]
HTTP アップロードに失敗したときのリトライ間隔と時間を設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210
43.6 HTTP アップロードで使用するプロキシサーバーの設定
[書式 ]
http upload  proxy  proxy  [port]
no http upload  proxy  [..]
[設定値及び初期値 ]
•proxy
• [設定値 ] : プロキシサーバー
• [初期値 ] : -
•port
• [設定値 ] : 1-65535586 | コマンドリファレンス  | HTTP アップロード機能

• [初期値 ] : 80
[説明 ]
HTTP アップロードで使用するプロキシサーバーを設定する。
[適用モデル ]
RTX1300, RTX1220, RTX1210
43.7 HTTP アップロードの実行
[書式 ]
http upload  go
[説明 ]
HTTP アップロードを実行する。
アップロードに失敗した場合、 http upload retry interval  コマンドの設定に基づいてリトライをする。
[ノート ]
alarm http upload  コマンドが  'on' の場合は、アップロードの成否に応じてアラーム音を鳴らす。
schedule at  コマンドで指定することができ、 startup を指定して起動時に実行させることもできる。
startup を指定した場合、起動直後は  HTTP サーバーへの経路が確立しておらずアップロードに失敗することがある。
こうした場合には http upload retry interval  コマンドの設定でリトライすることで対応できるようになる。
[適用モデル ]
RTX1300, RTX1220, RTX1210
43.8 HTTP アップロード機能に関連するアラーム音を鳴らすか否かの設定
[書式 ]
alarm  http upload  switch
no alarm  http upload  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 鳴らす
off 鳴らさない
• [初期値 ] : on
[説明 ]
HTTP アップロード機能に関連するアラーム音を鳴らすか否かを選択する。
[適用モデル ]
RTX1300, RTX1220, RTX1210コマンドリファレンス  | HTTP アップロード機能  | 587

