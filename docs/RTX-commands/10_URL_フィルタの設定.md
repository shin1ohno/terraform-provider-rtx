# 第10章: URL フィルタの設定

> 元PDFページ: 206-217

---

第 10 章
URL フィルタの設定
10.1 フィルタ定義の設定
[書式 ]
url filter  id kind keyword  [src_addr [/mask ]]
no url filter  id
[設定値及び初期値 ]
•id
• [設定値 ] :
設定値 説明
1..21474836 フィルタ番号
• [初期値 ] : -
•kind
• [設定値 ] :
設定値 説明
pass, pass-nolog 一致すれば通す  ( ログに記録しない  )
pass-log 一致すれば通す  ( ログに記録する  )
reject, reject-log 一致すれば破棄する  ( ログに記録する  )
reject-nolog 一致すれば破棄する  ( ログに記録しない  )
• [初期値 ] : -
•keyword
• [設定値 ] :
設定値 説明
任意の文字列 フィルタリングする  URL の全部もしくは一部  ( 半角
255 文字以内  )
* すべての  URL に対応
• [初期値 ] : -
•src_addr  : IP パケットの始点  IP アドレス
• [設定値 ] :
設定値 説明
任意の  IPv4 アドレス 1 個の  IPv4 アドレス
範囲指定 間に  - ( ハイフン  ) を挟んだ  2 つの  IP アドレス、 - を
後ろにつけた  IP アドレス、または  - を前につけた  IP
アドレス  ( 範囲指定  )
* すべての  IP アドレスに対応
省略 省略時は  * と同じ
• [初期値 ] : -
•mask
• [設定値 ] : ネットマスク長  ( src_addr  がネットワークアドレスの場合のみ指定可  )
• [初期値 ] : -206 | コマンドリファレンス  | URL フィルタの設定

[説明 ]
URL によるフィルタを設定する。本コマンドで設定されたフィルタは、  url interface  filter  コマンドで用いられる。
指定されたキーワードに、大文字のアルファベットが含まれる場合、それらを小文字に変換して保存する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
10.2 URL フィルタのインタフェースへの適用
[書式 ]
url interface  filter  dir list
url pp filter  dir list
url tunnel  filter  dir list
no url interface  filter  dir
no url pp filter  dir
no url tunnel  filter  dir
[設定値及び初期値 ]
•interface
• [設定値 ] : LAN インタフェース名、 WAN インタフェース名
• [初期値 ] : -
•dir
• [設定値 ] :
設定値 説明
in 入力方向の  HTTP コネクションをフィルタリングす
る
out 出力方向の  HTTP コネクションをフィルタリングす
る
• [初期値 ] : -
•list
• [設定値 ] : 空白で区切られた  URL フィルタ番号の並び  (512 個以内 ...RTX5000/RTX3500, 128 個以内 ...他の機
種 )
• [初期値 ] : -
[説明 ]
url filter  コマンドで設定したフィルタを組み合わせて、インタフェースで送受信する  HTTP パケットの  URL によっ
て制限を行う。
設定できるフィルタの数は、 RTX5000 、RTX3500 では  512 個以内、他の機種では  128 個以内、またはコマンドライ
ン文字列長  (4095 文字  ) で入力できる範囲内である。
指定されたすべてのフィルタにマッチしないパケットは破棄される。
[ノート ]
WAN インタフェースは  Rev.14.01 系以降のファームウェアで指定可能。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
10.3 URL フィルタでチェックを行う  HTTP のポート番号の設定
[書式 ]
url filter  port list
no url filter  port
[設定値及び初期値 ]
•listコマンドリファレンス  | URL フィルタの設定  | 207

• [設定値 ] : 空白で区切られたポート番号の並び  (4 個以内  )
• [初期値 ] : 80
[説明 ]
URL フィルタでチェックを行う  HTTP のポート番号を設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
10.4 URL フィルターを使用するか否かの設定
[書式 ]
url filter  use switch
no url filter  use
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on URL フィルターを使用する
off URL フィルターを使用しない
• [初期値 ] : on
[説明 ]
URL フィルターを使用するか否かを設定する。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
10.5 URL フィルタで破棄するパケットの送信元に  HTTP レスポンスを返す動作の設定
[書式 ]
url filter  reject  redirect
url filter  reject  redirect url
url filter  reject  off
no url filter  reject  [action ]
[設定値及び初期値 ]
• redirect : HTTP リダイレクトの  HTTP レスポンスを返し、ブロック画面へ転送する
• [初期値 ] : redirect ( RTX5000 、RTX3510 、RTX3500 以外の場合  )
• off : HTTP レスポンスは返さずに、 TCP RST によって  TCP セッションを終了する
• [初期値 ] : off ( RTX5000 、RTX3510 、RTX3500 の場合  )
•url
• [設定値 ] : リダイレクトする  URL(http:// または  https:// で始まる文字列で、半角  255 文字以内  )
• [初期値 ] : -
•action
• [設定値 ] :
• redirect
• off
• [初期値 ] : -
[説明 ]
URL フィルタで破棄するパケットの送信元に  HTTP レスポンスを返す動作を設定する。
ブロック画面には、一致したキーワードまたは、アクセスを遮断した理由を表示する。
url を指定した場合、 実際にリダイレクトするときには指定した  url の後ろに  "?" に続けて以下の内容のクエリを付加
する。208 | コマンドリファレンス  | URL フィルタの設定

•アクセスを遮断した  URL
•マッチしたフィルタに設定されているキーワード
url に http:// または  https:// で始まる文字列以外を設定することはできない。
[ノート ]
HTTP サーバー機能に対応した機種では、 redirect を設定して  Web ブラウザにブロック画面を表示する場合、 httpd
service  on の設定が必要である。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
10.6 フィルタにマッチした際にログを出力するか否かの設定
[書式 ]
url filter  log switch
no url filter  log
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on フィルタにマッチした際にログを出力する
off フィルタにマッチした際にログを出力しない
• [初期値 ] : on
[説明 ]
フィルタにマッチした際にログを出力するか否かを設定する。
[ノート ]
on を設定した場合でも、 url filter  コマンドで  kind に pass、pass-nolog 、または  reject-nolog を指定したフィルタにマ
ッチした場合はログを出力しない。
[適用モデル ]
RTX5000, RTX3510, RTX3500, RTX1300, RTX1220, RTX1210, RTX840, RTX830
10.7 利用するデータベースの選択
[書式 ]
url filter  external-database  use [reputation  reputation_name ] [category  category_name ]
no url filter  external-database  use
[設定値及び初期値 ]
•reputation_name  : Web レピュテーション機能で使用するデータベースの選択
• [設定値 ] :
設定値 説明
off Web レピュテーション機能を使用しない
trendmicro トレンドマイクロ株式会社のデータベースを使用す
る
• [初期値 ] : off
•category_name  : カテゴリーチェック機能で使用するデータベースの選択
• [設定値 ] :
設定値 説明
off カテゴリーチェック機能を使用しない
digitalarts2 デジタルアーツ株式会社のデータベースを使用するコマンドリファレンス  | URL フィルタの設定  | 209

設定値 説明
netstar ネットスター株式会社のデータベースを使用する
trendmicro トレンドマイクロ株式会社のデータベースを使用す
る
• [初期値 ] : off
[説明 ]
外部データベース参照型  URL フィルターの各評価機能で利用するデータベースを選択する。
データベースを利用するためには、 URL フィルタリングサービス事業者と契約を行う必要がある。
[ノート ]
ネットスター株式会社のデータベースは 2024年3月31日販売終了予定。詳しくはアルプスシステムインテグレー
ション株式会社にお問合せください。
[適用モデル ]
RTX5000, RTX3500
10.8 データベースを持つサーバーアドレスの設定
[書式 ]
url filter  external-database  server  address  port
no url filter  external-database  server
[設定値及び初期値 ]
•address  : サーバーのアドレス
• [設定値 ] :
• 1個の IPv4アドレス
• FQDN
• [初期値 ] : -
•port
• [設定値 ] : ポート番号 (1..65535)
• [初期値 ] : -
[説明 ]
外部データベースを持つサーバーのアドレス、およびデータベースにアクセスするためのポート番号を設定する。
デジタルアーツ株式会社のデータベースを使用する場合にのみ有効である。
[適用モデル ]
RTX5000, RTX3500
10.9 Proxy サーバーの設定
[書式 ]
url filter  external-database  proxy  server  address  [port]
no url filter  external-database  proxy  server
[設定値及び初期値 ]
•address  : Proxyサーバーのアドレス
• [設定値 ] :
• 1個の IPv4アドレス
• FQDN
• [初期値 ] : -
•port210 | コマンドリファレンス  | URL フィルタの設定

• [設定値 ] :
•ポート番号 (1..65535)
•省略時は 80
• [初期値 ] : -
[説明 ]
外部データベースを持つサーバーにアクセスする時に使用する Proxyサーバーのアドレス、ポート番号を設定す
る。
[適用モデル ]
RTX5000, RTX3500
10.10 チェックするカテゴリーの設定
[書式 ]
url filter  external-database  category  num kind category_list  [src_addr [/mask ]]
no url filter  external-database  category  num
[設定値及び初期値 ]
•num
• [設定値 ] : カテゴリーリスト番号 (1..21474836)
• [初期値 ] : -
•kind
• [設定値 ] :
設定値 説明
pass 一致すれば通す  ( ログに記録しない  )
pass-log 一致すれば通す  ( ログに記録する  )
pass-nolog 一致すれば通す  ( ログに記録しない  )
reject 一致すれば破棄する  ( ログに記録する  )
reject-log 一致すれば破棄する  ( ログに記録する  )
reject-nolog 一致すれば破棄する  ( ログに記録しない  )
• [初期値 ] : -
•category_list
• [設定値 ] :
•カテゴリー番号をコンマ (,)で区切った並び
• * ... すべててのカテゴリー番号
• [初期値 ] : -
•src_addr  : IPパケットの始点 IPアドレス
• [設定値 ] :
• IPv4 アドレス
•間にハイフン (-)を挟んだ 2つの上項目、 -を前につけた上項目、 -を後ろにつけた上項目、これらは範囲を
指定する
•上項目をコンマ (,)で区切った並び
• * ... すべての IPアドレス
•省略時は  * と同じ
• [初期値 ] : -
•mask
• [設定値 ] : ネットマスク長 (src_addr がネットワークアドレスの場合のみ指定可 )
• [初期値 ] : -コマンドリファレンス  | URL フィルタの設定  | 211

[説明 ]
URL フィルターでチェックするデータベースのカテゴリーを設定する。本コマンドで設定されたフィルターは、 url
interface  filter  コマンドで用いられる。
どのカテゴリーにも該当しない  URL は、 category_list  で * を指定した場合の設定が適用される。
指定できるカテゴリー番号は、使用する  URL フィルタリングサービス事業者により異なる。
[適用モデル ]
RTX5000, RTX3500
10.11 外部データベースへのアクセスに失敗したときにパケットを破棄するか否かの設定
[書式 ]
url filter  external-database  access  failure  type
no url filter  external-database  access  failure
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
pass パケットを通す
reject パケットを破棄する
• [初期値 ] : pass
[説明 ]
外部データベースへのアクセスに失敗したとき、パケットを破棄するか否かを設定する。
設定誤りによりデータベースを持つサーバーへアクセスできない、またはサーバーから応答がないなどの理由でサ
ーバーから正常な応答が得られなかった場合に、本コマンドの設定にしたがってパケットが処理される。
[適用モデル ]
RTX5000, RTX3500
10.12 URL フィルターで破棄するパケットの送信元に HTTPレスポンスを返す動作の設定
[書式 ]
url filter  external-database  reject  redirect  [url]
url filter  external-database  reject  redirect  url
url filter  external-database  reject  redirect  off
no url filter  external-database  reject
[設定値及び初期値 ]
•redirect  : HTTPリダイレクトの HTTPレスポンスを返し、ブロック画面へ転送する
• [初期値 ] : redirect (RTX5000 、RTX3500 以外の場合 )
•off : HTTPレスポンスは返さずに、 TCP RST によって TCPセッションを終了する
• [初期値 ] : off (RTX5000 、RTX3500 の場合 )
•url
• [設定値 ] : リダイレクトする URL (http:// または  https:// で始まる文字列で、半角 255文字以内 )
• [初期値 ] : -
[説明 ]
URLフィルターで破棄するパケットの送信元に HTTPレスポンスを返す動作を設定する。
URLを指定した場合、実際にリダイレクトするときには指定した URLの後ろに "?"に続けて以下の内容のクエリを
付加する。212 | コマンドリファレンス  | URL フィルタの設定

•使用している URLフィルタリング事業者名
•該当したセキュリティーレベル番号、カテゴリー番号、もしくはエラー文字列
•アクセスを遮断した URL
URLに http:// または  https:// で始まる文字列以外を設定することはできない。
[ノート ]
HTTPサーバー機能に対応した機種では、 redirectを設定して Webブラウザにブロック画面を表示する場合、 httpd
service  onの設定が必要である。
HTTPサーバー機能に対応していない機種で redirectを指定する場合、 urlを省略することはできない。
[適用モデル ]
RTX5000, RTX3500
10.13 IPアドレスを直接指定した URLへのアクセスを許可するか否かの設定
[書式 ]
url filter  external-database  ipaddress  access  type
no url filter  external-database  ipaddress  access
[設定値及び初期値 ]
•type
• [設定値 ] :
設定値 説明
pass パケットを通す
reject パケットを破棄する
• [初期値 ] : pass
[説明 ]
http://(IP アドレス )/XXXX のように、 IPアドレスを直接指定した URLへのアクセスを許可するか否かを設定する。
[適用モデル ]
RTX5000, RTX3500
10.14 指定した拡張子の URLを評価するか否かの設定
[書式 ]
url filter  external-database  lookup  specified  extension  switch
no url filter  external-database  lookup  specified  extension  [switch ]
[設定値及び初期値 ]
•switch
• [設定値 ] :
設定値 説明
on 指定した拡張子の URLを評価する
off 指定した拡張子の URLを評価しないで、通過させる
• [初期値 ] : off
[説明 ]
指定した拡張子の URLについて評価するか否かを設定する。評価を行わない場合、その URLへのリクエストを通
過させる。
初期設定として、以下の拡張子が登録されている。コマンドリファレンス  | URL フィルタの設定  | 213

jpg, gif, ico, png, bmp, jpeg, tif, tiff, swf, wav, wmv, wma, mp3, mpg, mpeg, mp4, asx, asf, wax, wvx, mov
url filter external-database lookup specified extension list コマンドで上記拡張子のリストに拡張子を追加または削除
することができる。
[ノート ]
使用する外部データベースに対応するリビジョンは次のとおり。
•ネットスター株式会社
RTX3500/RTX5000 Rev14.00.08 以降のファームウェアで有効。
※ ネットスター株式会社のデータベースは 2024年3月31日販売終了予定。詳しくはアルプスシステムインテ
グレーション株式会社にお問合せください。
[適用モデル ]
RTX5000, RTX3500
10.15 評価しない URLの拡張子の設定
[書式 ]
url filter  external-database  lookup  specified  extension  list [+|-]extension  [..]
no url filter  external-database  lookup  specified  extension  list [..]
[設定値及び初期値 ]
•extension
• [設定値 ] : 拡張子 (半角 4文字以内、 64個以内 )
• [初期値 ] : -
[説明 ]
url filter external-database lookup specified extension コマンドが offの設定の場合に、 評価せずリクエストを通過させ
るURLの拡張子を設定する。
初期設定として、以下の拡張子が登録されており、このリストへの追加、削除する形で拡張子を設定する。
jpg, gif, ico, png, bmp, jpeg, tif, tiff, swf, wav, wmv, wma, mp3, mpg, mpeg, mp4, asx, asf, wax, wvx, mov
extension の前に +を置くか、あるいは何も置かない場合には上記初期設定のリストに extension を追加する。
extension の前に -を置く場合には上記初期設定のリストから extension を削除する。
[ノート ]
使用する外部データベースに対応するリビジョンは次のとおり。
•ネットスター株式会社
RTX3500/RTX5000 Rev14.00.08 以降のファームウェアで有効。
※ ネットスター株式会社のデータベースは 2024年3月31日販売終了予定。詳しくはアルプスシステムインテ
グレーション株式会社にお問合せください。
[適用モデル ]
RTX5000, RTX3500
10.16 フィルターにマッチした際にログを出力するか否かの設定
[書式 ]
url filter  external-database  log switch
no url filter  external-database  log
[設定値及び初期値 ]
•switch
• [設定値 ] :214 | コマンドリファレンス  | URL フィルタの設定

設定値 説明
on フィルターにマッチした際にログを出力する
off フィルターにマッチした際にログを出力しない
• [初期値 ] : on
[説明 ]
フィルターにマッチした際にログを出力するか否かを設定する。
[ノート ]
onを設定した場合でも、 url filter コマンドで kindにpass、pass-nolog 、または reject-nolog を指定したフィルターに
マッチした場合はログを出力しない。
[適用モデル ]
RTX5000, RTX3500
10.17 シリアル IDを登録する URLの設定
[書式 ]
url filter  external-database  register  url url
no url filter  external-database  register  url
[設定値及び初期値 ]
•url
• [設定値 ] : シリアル IDを登録する URL(半角 255文字以内 )
• [初期値 ] : https://ars2s.daj.co.jp/register/add.php
[説明 ]
シリアル IDを登録する URLを設定する。
デジタルアーツ株式会社のデータベースを使用する場合にのみ有効である。
[適用モデル ]
RTX5000, RTX3500
10.18 データベースへアクセスするためのシリアル IDの設定
[書式 ]
url filter  external-database  id name  [id]
no url filter  external-database  id name
[設定値及び初期値 ]
•name
• [設定値 ] :
設定値 説明
digitalarts デジタルアーツ株式会社のデータベースへアクセス
するためのシリアル IDを設定する
trendmicro トレンドマクロ株式会社のデータベースへアクセス
するためのアクティベーションコードを設定する
• [初期値 ] : -
•id
• [設定値 ] : シリアル ID(半角 255文字以内 )
• [初期値 ] : -コマンドリファレンス  | URL フィルタの設定  | 215

[説明 ]
各サービス事業者のデータベースへアクセスするためのシリアル IDを設定する。
[適用モデル ]
RTX5000, RTX3500
10.19 URL フィルタリングサービス事業者にシリアル IDの登録
[書式 ]
url filter  external-database  id activate  go [database ]
[設定値及び初期値 ]
•database
• [設定値 ] :
設定値 説明
reputation Webレピュテーションデータベースのサービス事業
者にシリアル IDを登録する
category カテゴリデータベースのサービス事業者にシリアル
IDを登録する
• [初期値 ] : -
[説明 ]
url filter external-database use コマンドの設定に従い、 URLフィルタリングサービス事業者にシリアル IDを登録す
る。
database パラメーターを指定することで、特定のデータベースのサービス事業者との契約状況のみを確認する。
[ノート ]
本コマンドを実行する前に、 url filter external-database use コマンドで、使用するデータベースを設定し、 url filter
external-database id  コマンドで、シリアル IDを設定する必要がある。
トレンドマイクロ株式会社のデータベースを使用する場合にのみ有効である。
[適用モデル ]
RTX5000, RTX3500
10.20 URL フィルタリングサービス事業者との契約状況の確認
[書式 ]
url filter  external-database  id check  go [database ]
[設定値及び初期値 ]
•database
• [設定値 ] :
設定値 説明
reputation Webレピュテーションデータベースのサービス事業
者との契約状況を確認する
category カテゴリデータベースのサービス事業者との契約状
況を確認する
• [初期値 ] : -
[説明 ]
url filter external-database use コマンドの設定に従い、 URLフィルタリングサービス事業者との契約状況を確認す
る。216 | コマンドリファレンス  | URL フィルタの設定

database パラメーターを指定することで、特定のデータベースのサービス事業者との契約状況のみを確認する。ま
た、 database パラメーターを省略し、且つ複数のサービス事業者のデータベースを使用している場合は、それぞれ
の契約状況を確認する。
[ノート ]
本コマンドを実行する前に、 url filter external-database use コマンドで、使用するデータベースを設定する必要があ
る。
[適用モデル ]
RTX5000, RTX3500
10.21 ユーザー認証に失敗した場合の再送間隔と回数の設定
[書式 ]
url filter  external-database  auth  retry  interval  [retry ]
no url filter  external-database  auth  retry
[設定値及び初期値 ]
•interval
• [設定値 ] :
設定値 説明
60 .. 300 再送間隔
auto 自動
off 再送しない
• [初期値 ] : auto
•retry
• [設定値 ] :
設定値 説明
1 .. 50 再送回数
• [初期値 ] : 10
[説明 ]
外部データベース参照型 URLフィルターでユーザー認証の自動実行に失敗した場合に、再度ユーザー認証を実行す
る間隔と回数を設定する。
intervalにautoを設定した時に、ユーザー認証に失敗した場合には 30秒から 90秒の時間をおいて再度ユーザー認証
を行う。それにも失敗した場合には、その後 60秒間隔でユーザー認証を試みる。
intervalにoffを設定した時には、ユーザー認証に失敗した場合でも再送は行わない。
retryはintervalにoff以外を設定した場合に指定できる。
[ノート ]
url filter external-database id check go コマンドで、手動でユーザー認証を実行した場合には、本コマンドでの設定に
かかわらずユーザー認証の再送は行われない。
ユーザー認証に失敗してから指定した時間までの間にユーザー認証を手動実行した場合には、 その後の intervalで指
定した再送間隔でのユーザー認証は行わない。
[適用モデル ]
RTX5000, RTX3500コマンドリファレンス  | URL フィルタの設定  | 217

