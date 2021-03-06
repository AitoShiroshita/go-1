// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	fmtパッケージは、C言語のprintfおよびscanfに似た機能を持つ、
	フォーマットのためのI/Oを実装します。フォーマットに用いる「動詞」は
	C言語から派生されていますが、より簡素なものになっています。


	プリント

	動詞：

	一般：
		%v	デフォルトフォーマットの値。
			構造体をプリントするときは、プラスフラグ（%+v）にフィールド名を付加します。
		%#v	値のGo言語構文表現。
		%T	値の型のGo言語構文表現。
		%%	パーセントリテラル記号で、値を利用しません。

	ブール値：
		%t	trueもしくはfalseの値
	整数値：
		%b	2進数
		%c	Unicodeのコードポイントに対応する文字
		%d	10進数
		%o	8進数
		%q	GO言語構文でエスケープしシングルクオートで囲んだ文字リテラル
		%x	a-fの小文字を用いた16進数
		%X	A-Fの大文字を用いた16進数
		%U	Unicodeフォーマット：U+1234、"U+%04X"と同様
	浮動小数点数値および複素数の構成要素：
		%b	2のべき乗を用いた小数を使わない指数表現
			'b'フォーマットを用いたstrconv.FormatFloatと同じ表記となる
			例：-123456p-78
		%e	指数表現、例：-1234.456e+78
		%E	指数表現、例：-1234.456E+78
		%f	指数を用いない小数点表現、例：123.456
		%g	%e、%fのうち、より簡素な表現
		%G	%E、%fのうち、より簡素な表現
	stringおよびbyte配列のslice：
		%s	解釈されないstringのbyte配列もしくはslice
		%q	Go言語構文で安全にエスケープ処理しダブルクオーテーションで囲ったstring
		%x	16進数、小文字、1バイト2文字で表現
		%X	16進数、大文字、1バイト2文字で表現
	ポインタ：
		%p	0xで始まる16進数表現

	'u'フラグはありません。整数は符号なしの型の場合は符号なしでプリントされます。
	同様に、オペランドのサイズ（int8、int64等）を指定する必要はありません。

	有効幅と精度はフォーマットを調整しますが、Unicodeコードポイントの単位となります。
	（これは単位がバイト数となるC言語のprintfとは異なります。）
	どちらか一方もしくは両方のフラグを'*'文字に置き換えることができます。
	それらの実際の値は次のオペランドから得られることになりますが、必ずint型である必要があります。

	数値については、有効幅がフィールドの最小幅、精度が小数点の後の桁数を可能であれば設定します。
	ただし、%g/%Gの場合は例外で、精度として総桁数を設定します。例えば、123.45を%6.2fと
	フォーマットした場合には123.45、%.4gとフォーマットした場合には123.5とプリントします。
	%eおよび%fのデフォルト精度は6、%gのデフォルト精度は値を一意に識別するために
	必要な最小の桁数となります。

	ほとんどの値について、有効幅は表示する際の最小の文字数となり、必要であれば
	フォーマットした形式をスペースでパディングします。stringについては、
	精度は表示する際の最大の文字数となり、必要であれば切り詰めを行ないます。

	その他のフラグ：
		+	数値に対し常に符号をプリントします。
			%g（つまり%+g）についてはASCIIのみでの出力を保証します。
		-	左側ではなく、右側にスペースでパディングします（フィールドを左揃えにします）。
		#	代替フォーマット。8進数（%#o）の頭に0、16進数（%#ｘ）に0x、
			16進数（%#X）に0Xを付加し、%p (つまり%#p)から0xを除きます。
			%gについてはstrconv.CanBackquoteがtrueを返す場合はバッククオートを
			を用いた生のstringをプリントします。
			characterが%U（つまり%#U）で表示できる場合は、例えばU+0078を'x'と出力します。
		' '	（スペース）数値（たとえば% d）については符号を表示しない場合に符号の分のスペースを残します。
			16進数（たとえば% x、% X）についてはstringやsliceをプリントするbyteの間にスペースを挿入します。
		0	先頭をスペースではなく0でパディングします。
			数値については、符号の後にパディングします。

	動詞に対し利用できないフラグを指定した場合には無視されます。
	たとえば、10進数には代替フォーマットがないため、%#dと%dはまったく同じ動作となります。

	それぞれのPrintf系の関数に対して、フォーマットを持たないPrint関数も存在します。
	Print関数はすべてのオペランドに対し%vでプリントさせているのと同等に動作します。
	もう一つの類似関数であるPrintlnはオペランドの間に空白を挿入し、最後に改行文字を追加します。

	動詞にかかわらず、オペランドがインタフェースの値であった場合には、
	そのインタフェース自身ではなく内部の具体的な値が利用されます。
	つまり、
		var i interface{} = 23
		fmt.Printf("%v\n", i)
	は23をプリントします。
		
	オペランドがFormatterインタフェースを実装している場合は、
	インタフェースでフォーマットを微調整することができます。

	フォーマット（Printlnなどで暗黙的に%vを用いる場合を含む）が
	正しいstring（%s、%q、%v、%x、%X）の場合、下記の2ルールが追加で適用されます。

	1. オペランドがerrorインタフェースを実装している場合、
	Errorメソッドはstringを変換するために用いられ、
	動詞がある場合はそのstringが動詞に従ってフォーマットされます。

	2. オペランドがメソッドString() stringを実装している場合、
	そのメソッドはオブジェクトをstringに変換するために用いられ、
	動詞がある場合はそのstringが動詞に従ってフォーマットされます。

	以下のような再帰呼び出しを避けるために、
		type X string
		func (x X) String() string { return Sprintf("<%s>", x) }
	値を変換して下さい。
		func (x X) String() string { return Sprintf("<%s>", string(x)) }

	明示的な引数のインデックス：

	Printf、Sprintf、Fprintfでは、デフォルトでは関数呼び出しの中で
	それぞれのフォーマット動詞の次に続く引数をフォーマットするように動作します。
	しかし、動詞のすぐ前に[n]表記を挿入すると、代わりにn番目の引数が
	フォーマットされるようになります。有効幅および精度については
	'*'の前に同じ記法で記述すると、対応する順番の引数で指定することができます。
	一度このカギカッコを用いた[n]記法を記述すると、そのあとは明示的に指定しなければ
	順にn+1、n+2...番目の引数が用いられます。

	例：
		fmt.Sprintf("%[2]d %[1]d\n", 11, 22)
	は "22, 11" という結果となります。一方、
		fmt.Sprintf("%[2]d %[1]d\n", 11, 22)
	は下記と同等で
		fmt.Sprintf("%6.2f", 12.0),
	" 12.00" という結果となります。明示的なインデックス指定は続く動詞を
	制御できるため、この記法により繰り返す最初の引数のインデックスに
	リセットすることで同じ値を複数回表示することが可能です。
		fmt.Sprintf("%d %d %#[1]x %#x", 16, 17)
	の結果は "16 17 0x10 0x11" となります。

	フォーマットエラー：

	%dにstringを指定するなど、動詞に対し不正な引数が与えられると
	表示されたstringは下記の例のように問題のある表現となります。

		型の誤りや不明な動詞：%!verb(type=value)
			Printf("%d", hi):          %!d(string=hi)
		引数が多すぎる：%!(EXTRA type=value)
			Printf("hi", "guys"):      hi%!(EXTRA string=guys)
		引数が少なすぎる：%!verb(MISSING)
			Printf("hi%d"):            hi %!d(MISSING)
		有効幅や精度に対してint以外を指定：%!(BADWIDTH) or %!(BADPREC)
			Printf("%*s", 4.5, "hi"):  %!(BADWIDTH)hi
			Printf("%.*s", 4.5, "hi"): %!(BADPREC)hi
		不正な引数のインデックスや引数のインデックスの不正な利用：%!(BADINDEX)
			Printf("%*[2]d", 7):       %!d(BADINDEX)
			Printf("%.[2]d", 7):       %!d(BADINDEX)

	すべてのエラーは"%!"の文字列で始まり、その後に1文字（動詞）が入る場合があり、
	最後はカッコ表記で終わります。

	ErrorもしくはStringメソッドがプリント処理で呼び出された時に
	panicを引き起こした場合には、fmtパッケージはpanicからの
	エラーメッセージをfmtパッケージでの表記に準じて再フォーマットします。
	例えば、Stringメソッドがpanic("bad")を呼び出した場合、
	フォーマットしたメッセージは以下のような結果になります。
		%!s(PANIC=bad)

	エラーが起きた場合には%!sがprintでの動詞を表すために用いられます。

	スキャン

	値を取得するためにフォーマットしたテキストをスキャンする一連の類似関数を提供します。
	Scan、Scanf、Scanlnはos.Stdinから読み、FScan、Fscanf、Fscanlnは
	指定されたio.Readerから読み、Sscan、Sscanf、Sscanlnは引数に指定した
	stringから読みます。Scanln、Fscanln、Sscanlnは改行でスキャンを中止します。
	スキャンされる項目は必ず改行を含んでいる必要があります。
	Scanf、Fscanf、Sscanfでは入力された改行はフォーマットの改行と一致している
	必要があります。その他の処理では改行はスペースとして扱われます。

	Scanf、Fscanf、SscanfはPrintfと同様にstringで記述されたフォーマットに
	したがって引数の内容を分析します。例えば、%xは整数を16進数としてスキャンし、
	%vは値に対しデフォルト表現のフォーマットでスキャンします。

	以下の点を除き、Printfと同様のフォーマットの振る舞いをします。

		%pは実装されていません
		%Tは実装されていません
		%e %E %f %F %g %Gはすべて同等でどのような浮動小数点数や複素数でも読み込みます
		%s、%vはスペースで区切られたトークンをスキャンします
		#フラグ、+フラグは実装されていません

	おなじみの基数設定のプレフィックスである0（8進数）、0x（16進数）は
	%v動詞を用いるか、全くフォーマットを用いない場合にスキャンすることができます。

	有効幅は入力されたテキストから解釈できますが（%5sは最大5ルーンの入力を
	スキャンすることを意味します）、精度を指定する記法はありません。
	（%5.2fは指定できません。%5fのみ指定できます。）

	フォーマットを用いてスキャンする際、連続したスペース（改行を除く）は
	フォーマット・入力のいずれにおいても1つのスペースと同等に扱われます。
	この条件で、stringで記述されたフォーマットのテキストは入力されたテキストと
	一致しなければなりません。そうでなければ、スキャンできた引数の数を示した
	関数の戻り値を返して、スキャンを中止します。

	これらすべてのスキャン関数において、直後に改行が続くキャリッジリターン
	は通常の改行として扱われます。（\r\nは\nと同じ意味です）

	また、これらすべてのスキャン関数において、オペランドがScanメソッドを
	実装している場合（つまり、Scannerインタフェースを実装している場合）、
	そのオペランドについては該当のメソッドがテキストを読み込むのに利用されます。
	また、スキャンされた引数の数が与えられた引数の数より少なかった場合には
	エラーが返されます。

	スキャンされた引数はすべて基本型かScannerインタフェースの実装への
	ポインタでなければなりません。

	注：Fscan等は入力が返された後で1文字（ルーン）を読むことができます。
	つまり、スキャン処理を繰り返し呼ぶ場合にはいくつかの入力がスキップされます。
	これは通常は入力値の間にスペースがない場合のみに起こる問題です。
	Fscanに与えられたreaderがReadRuneを実装している場合、そのメソッドは
	文字を読むために利用されます。readerがさらにUnreadRuneを実装している場合には
	そのメソッドは文字情報を保持し、後に続く呼び出しでデータが失われないようにします。
	readerにReadRuneメソッドおよびUnreadRuneメソッドを付け実装するのが難しい場合には
	bufio.NewReaderを利用して下さい。

	本ドキュメントは以下のドキュメントを翻訳しています: https://code.google.com/p/go/source/browse/src/pkg/fmt/doc.go?r=09a1cf0d94b2b6cf672f8dcd04c6962e0916bc4e
*/
package fmt
