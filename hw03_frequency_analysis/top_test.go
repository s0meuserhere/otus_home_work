package hw03frequencyanalysis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = true

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "only whitespace",
			input: "   \t\n  ",
			want:  []string{},
		},
		{
			name:  "single word",
			input: "hello",
			want:  []string{"hello"},
		},
		{
			name:  "less than 10 unique words",
			input: "a b a b",
			want:  []string{"a", "b"},
		},
		{
			name:  "frequency over lexicographic order",
			input: "bbb zzz zzz zzz aaa aaa",
			want:  []string{"zzz", "aaa", "bbb"},
		},
		{
			name:  "lexicographic order",
			input: "bbb aaa ccc aaa bbb ccc",
			want:  []string{"aaa", "bbb", "ccc"},
		},
		{
			name:  "more than 10 words with same frequency",
			input: "a b c d e f g h i j k l m n o",
			want:  []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{
			name:  "exactly 10 unique words",
			input: "a b c d e f g h i j",
			want:  []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{
			name:  "case insensitive",
			input: "Нога нога НОГА",
			want:  []string{"нога"},
		},
		{
			name:  "trim punctuation",
			input: "нога! 'нога' нога, \"нога\"",
			want:  []string{"нога"},
		},
		{
			name:  "keep punctuation inside word",
			input: "dog,cat dog...cat dogcat",
			want:  []string{"dog,cat", "dog...cat", "dogcat"},
		},
		{
			name:  "different word forms",
			input: "нога ногу ноги",
			want:  []string{"нога", "ноги", "ногу"},
		},
		{
			name:  "dash inside word",
			input: "Винни-Пух винни-пух",
			want:  []string{"винни-пух"},
		},
		{
			name:  "dash rules",
			input: "- ------- какой-то какойто -",
			want:  []string{"-------", "какой-то", "какойто"},
		},
		{
			name:  "readme example",
			input: "cat and dog, one dog,two cats and one man",
			want:  []string{"and", "one", "cat", "cats", "dog", "dog,two", "man"},
		},
		{
			name:  "words are different",
			input: "какой-то какойто",
			want:  []string{"какой-то", "какойто"},
		},
		{
			name:  "top word plus lexicographic tail",
			input: "word word word word word a b c d e f g h i j k",
			want:  []string{"word", "a", "b", "c", "d", "e", "f", "g", "h", "i"},
		},
		{
			name:  "single dash is not a word",
			input: "- - -",
			want:  []string{},
		},
		{
			name:  "dash sequence is a word",
			input: "-------",
			want:  []string{"-------"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Top10(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}
