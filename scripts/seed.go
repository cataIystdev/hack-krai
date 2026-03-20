// Скрипт seed.go очищает таблицу locations и заливает 20 тестовых локаций
// с реальными координатами Краснодарского края.
// Запуск: go run scripts/seed.go
//
// Предварительные условия:
// - PostgreSQL с PostGIS запущен (docker compose up -d postgres)
// - Миграции 001 и 002 применены
// - Существует пользователь-хост для привязки owner_id
//
// Скрипт автоматически создаёт пользователя-хоста seed_host@deepkrai.ru
// если его нет, и привязывает все локации к нему.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

// seedLocation — данные для вставки одной локации.
type seedLocation struct {
	Name             string
	DescriptionShort string
	DescriptionFull  string
	Category         string
	Tags             []string
	PricePerNight    int
	Capacity         int
	AccessLevel      string
	DensityLevel     string
	ChildFriendly    bool
	Address          string
	IsPublished      bool
	Latitude         float64
	Longitude        float64
	Slug             string
}

func main() {
	// Загрузка конфигурации из .env файла.
	viper.SetConfigFile("../.env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			log.Println("Предупреждение: .env файл не найден, используются переменные окружения")
		}
	}

	host := envOrDefault("POSTGRES_HOST", "localhost")
	port := envOrDefault("POSTGRES_PORT", "5432")
	user := envOrDefault("POSTGRES_USER", "deepkrai")
	password := envOrDefault("POSTGRES_PASSWORD", "deepkrai_secret_2026")
	db := envOrDefault("POSTGRES_DB", "deepkrai")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer pool.Close()

	log.Println("Подключение к PostgreSQL установлено")

	// Создание или получение пользователя-хоста для seed-данных.
	ownerID, err := ensureSeedHost(ctx, pool)
	if err != nil {
		log.Fatalf("Ошибка создания seed-хоста: %v", err)
	}
	log.Printf("Seed-хост ID: %s\n", ownerID)

	// Очистка таблицы locations.
	_, err = pool.Exec(ctx, "DELETE FROM locations")
	if err != nil {
		log.Fatalf("Ошибка очистки таблицы locations: %v", err)
	}
	log.Println("Таблица locations очищена")

	// Вставка 20 тестовых локаций.
	locations := getSeedLocations()
	for i, loc := range locations {
		_, err := pool.Exec(ctx, `
			INSERT INTO locations (
				owner_id, slug, name, description_short, description_full,
				category, tags, price_per_night, capacity,
				access_level, density_level, child_friendly,
				address, is_published,
				geo
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, $11, $12,
				$13, $14,
				ST_SetSRID(ST_MakePoint($15, $16), 4326)
			)`,
			ownerID, loc.Slug, loc.Name, loc.DescriptionShort, loc.DescriptionFull,
			loc.Category, loc.Tags, loc.PricePerNight, loc.Capacity,
			loc.AccessLevel, loc.DensityLevel, loc.ChildFriendly,
			loc.Address, loc.IsPublished,
			loc.Longitude, loc.Latitude,
		)
		if err != nil {
			log.Fatalf("Ошибка вставки локации %d (%s): %v", i+1, loc.Name, err)
		}
		log.Printf("[%d/20] %s (%.4f, %.4f)\n", i+1, loc.Name, loc.Latitude, loc.Longitude)
	}

	log.Println("Seed завершён: 20 локаций Краснодарского края успешно загружены")
}

// ensureSeedHost создаёт пользователя-хоста если его нет, возвращает ID.
func ensureSeedHost(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	var id string
	err := pool.QueryRow(ctx, "SELECT id FROM users WHERE email = 'seed_host@deepkrai.ru'").Scan(&id)
	if err == nil {
		return id, nil
	}

	// Создание хоста (пароль: bcrypt hash "seedpassword123").
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, display_name)
		VALUES ('seed_host@deepkrai.ru', '$2a$12$LJ3m4ys2Kq5yGZvADfQZ3OZwNPBiDp6hzKqCzVONwGDHfQKh0IKPa', 'host', 'Seed Host')
		ON CONFLICT (email) DO UPDATE SET role = 'host'
		RETURNING id
	`).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("ошибка создания seed-хоста: %w", err)
	}
	return id, nil
}

// envOrDefault возвращает значение переменной окружения или значение по умолчанию.
func envOrDefault(key, defaultVal string) string {
	if v := viper.GetString(key); v != "" {
		return v
	}
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getSeedLocations возвращает 20 тестовых локаций Краснодарского края.
func getSeedLocations() []seedLocation {
	return []seedLocation{
		{
			Name:             "Винодельня Абрау-Дюрсо",
			Slug:             "vinodelnya-abrau-dyurso",
			DescriptionShort: "Легендарная винодельня на берегу озера Абрау",
			DescriptionFull:  "Винодельня Абрау-Дюрсо -- старейшее производство игристых вин в России, основанное в 1870 году. Подземные тоннели протяженностью 5 км хранят миллионы бутылок шампанского. Дегустации, экскурсии и фирменный ресторан с видом на горное озеро.",
			Category:         "winery",
			Tags:             []string{"вино", "дегустация", "озеро", "экскурсия", "история"},
			PricePerNight:    0,
			Capacity:         50,
			AccessLevel:      "open",
			DensityLevel:     "red",
			ChildFriendly:    false,
			Address:          "Краснодарский край, пос. Абрау-Дюрсо",
			IsPublished:      true,
			Latitude:         44.6979,
			Longitude:        37.5949,
		},
		{
			Name:             "Винодельня Лефкадия",
			Slug:             "vinodelnya-lefkadiya",
			DescriptionShort: "Премиальная винодельня в долине Лефкадия",
			DescriptionFull:  "Долина Лефкадия -- современная винодельня с авторскими виноградниками на 80 гектарах. Собственная сыроварня, ресторан и дегустационные залы. Органические вина, получившие международные награды.",
			Category:         "winery",
			Tags:             []string{"вино", "сыр", "органика", "дегустация", "ресторан"},
			PricePerNight:    12000,
			Capacity:         20,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    true,
			Address:          "Краснодарский край, с. Молдаванское",
			IsPublished:      true,
			Latitude:         44.7230,
			Longitude:        37.7810,
		},
		{
			Name:             "Козья ферма дяди Вани",
			Slug:             "kozya-ferma-dyadi-vani",
			DescriptionShort: "Крафтовые сыры и горный мёд в тишине предгорий",
			DescriptionFull:  "Крафтовые сыры с благородной плесенью по авторской рецептуре. 12 альпийских коз на горном пастбище. Два уютных гостевых домика с видом на закат. Дегустация сыров и мёда, мастер-классы по сыроварению.",
			Category:         "farm",
			Tags:             []string{"сыр", "козы", "ночлег", "тишина", "горы", "мёд"},
			PricePerNight:    5000,
			Capacity:         4,
			AccessLevel:      "semi_open",
			DensityLevel:     "green",
			ChildFriendly:    true,
			Address:          "Краснодарский край, Хаджох (Каменномостский)",
			IsPublished:      true,
			Latitude:         44.2878,
			Longitude:        40.1763,
		},
		{
			Name:             "Плато Лаго-Наки",
			Slug:             "plato-lago-naki",
			DescriptionShort: "Альпийские луга на высоте 2000 метров",
			DescriptionFull:  "Плато Лаго-Наки -- одно из самых красивых мест Кавказа. Альпийские луга, карстовые пещеры, ледниковые озёра и панорамные виды на Главный Кавказский хребет. Треккинг, спелеотуризм и горнолыжные спуски зимой.",
			Category:         "trail",
			Tags:             []string{"горы", "треккинг", "природа", "пещеры", "панорама"},
			PricePerNight:    0,
			Capacity:         100,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    false,
			Address:          "Республика Адыгея, плато Лаго-Наки",
			IsPublished:      true,
			Latitude:         44.0580,
			Longitude:        39.9800,
		},
		{
			Name:             "Гостевой дом Горная тишина",
			Slug:             "gostevoy-dom-gornaya-tishina",
			DescriptionShort: "Уютный горный приют в станице Даховская",
			DescriptionFull:  "Гостевой дом в станице Даховская у подножия гор. Деревянные домики, камин, открытая терраса с видом на реку Белую. Идеальная база для походов в Хаджохскую теснину и на плато Лаго-Наки.",
			Category:         "guesthouse",
			Tags:             []string{"ночлег", "горы", "река", "камин", "тишина"},
			PricePerNight:    3500,
			Capacity:         8,
			AccessLevel:      "open",
			DensityLevel:     "green",
			ChildFriendly:    true,
			Address:          "Республика Адыгея, ст. Даховская",
			IsPublished:      true,
			Latitude:         44.2613,
			Longitude:        40.2048,
		},
		{
			Name:             "Гастро-маркет Море вкусов",
			Slug:             "gastro-market-more-vkusov",
			DescriptionShort: "Фермерские продукты и уличная кухня Кубани",
			DescriptionFull:  "Открытый гастрономический рынок с фермерскими продуктами Краснодарского края. Крафтовые сыры, домашнее вино, морепродукты, кубанские деликатесы. Мастер-классы по кубанской кухне каждую субботу.",
			Category:         "gastro",
			Tags:             []string{"еда", "фермерское", "рынок", "морепродукты", "мастер-класс"},
			PricePerNight:    0,
			Capacity:         200,
			AccessLevel:      "open",
			DensityLevel:     "red",
			ChildFriendly:    true,
			Address:          "г. Краснодар, ул. Красная",
			IsPublished:      true,
			Latitude:         45.0355,
			Longitude:        38.9753,
		},
		{
			Name:             "Скала Парус",
			Slug:             "skala-parus",
			DescriptionShort: "Природный памятник на побережье Геленджика",
			DescriptionFull:  "Скала Парус -- уникальный природный памятник высотой 25 метров на берегу Чёрного моря. Тонкая каменная стена, напоминающая парус, вырастает прямо из воды. Одинокий пляж, дикая природа и потрясающие закаты.",
			Category:         "nature",
			Tags:             []string{"скала", "море", "природа", "закат", "пляж"},
			PricePerNight:    0,
			Capacity:         30,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    true,
			Address:          "Краснодарский край, г. Геленджик",
			IsPublished:      true,
			Latitude:         44.4343,
			Longitude:        38.1793,
		},
		{
			Name:             "33 водопада",
			Slug:             "33-vodopada",
			DescriptionShort: "Каскад водопадов в Сочинском национальном парке",
			DescriptionFull:  "Каскад из 33 водопадов на ручье Джегош в Лазаревском районе. Высота самого большого -- 12 метров. Деревянные мостки ведут через тропический лес с самшитовыми рощами и лианами.",
			Category:         "nature",
			Tags:             []string{"водопады", "лес", "тропа", "природа", "купание"},
			PricePerNight:    0,
			Capacity:         50,
			AccessLevel:      "open",
			DensityLevel:     "red",
			ChildFriendly:    true,
			Address:          "Краснодарский край, Лазаревский район",
			IsPublished:      true,
			Latitude:         43.8436,
			Longitude:        39.5613,
		},
		{
			Name:             "Виноградники Мысхако",
			Slug:             "vinogradniki-myskhako",
			DescriptionShort: "Старейшие виноградники под Новороссийском",
			DescriptionFull:  "Виноградники Мысхако -- одно из старейших виноградарских хозяйств юга России. Подвалы для выдержки вин вырублены в скальной породе в XIX веке. Авторские туры по виноградникам с дегустацией.",
			Category:         "winery",
			Tags:             []string{"вино", "виноградники", "история", "дегустация"},
			PricePerNight:    0,
			Capacity:         30,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    false,
			Address:          "Краснодарский край, пос. Мысхако",
			IsPublished:      true,
			Latitude:         44.6571,
			Longitude:        37.7710,
		},
		{
			Name:             "Горячий Ключ: Дантово ущелье",
			Slug:             "goryachiy-klyuch-dantovo-ushchelye",
			DescriptionShort: "Таинственное ущелье и целебные источники",
			DescriptionFull:  "Дантово ущелье -- рукотворный каньон XIX века, прорубленный казаками в скале. Стены покрыты мхом и папоротниками. Рядом -- минеральные источники и курортный парк с вековыми деревьями.",
			Category:         "nature",
			Tags:             []string{"ущелье", "источники", "парк", "лечение", "история"},
			PricePerNight:    0,
			Capacity:         100,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    true,
			Address:          "Краснодарский край, г. Горячий Ключ",
			IsPublished:      true,
			Latitude:         44.6307,
			Longitude:        39.0956,
		},
		{
			Name:             "Глэмпинг Поляна у ручья",
			Slug:             "glemping-polyana-u-ruchya",
			DescriptionShort: "Палаточный лагерь с комфортом у горного ручья",
			DescriptionFull:  "Глэмпинг в предгорьях Кавказа: просторные палатки-купола с кроватями, электричеством и душем. Ручей для купания, костровая зона, йога на рассвете и звёздное небо без засветки.",
			Category:         "camping",
			Tags:             []string{"глэмпинг", "палатки", "горы", "звёзды", "ручей"},
			PricePerNight:    4500,
			Capacity:         12,
			AccessLevel:      "semi_open",
			DensityLevel:     "green",
			ChildFriendly:    true,
			Address:          "Краснодарский край, пос. Мезмай",
			IsPublished:      true,
			Latitude:         44.1990,
			Longitude:        39.9580,
		},
		{
			Name:             "Пасека деда Михалыча",
			Slug:             "paseka-deda-mikhalycha",
			DescriptionShort: "Горный мёд по старинным рецептам",
			DescriptionFull:  "Семейная пасека в горах уже четыре поколения. 50 ульев на альпийских лугах. Дегустация мёда: горный, каштановый, липовый, разнотравье. Мастер-класс по откачке мёда и экскурсия по пасеке.",
			Category:         "farm",
			Tags:             []string{"мёд", "пасека", "горы", "экскурсия", "мастер-класс"},
			PricePerNight:    0,
			Capacity:         10,
			AccessLevel:      "hidden",
			DensityLevel:     "green",
			ChildFriendly:    true,
			Address:          "Краснодарский край, ст. Азовская",
			IsPublished:      true,
			Latitude:         44.5500,
			Longitude:        38.5600,
		},
		{
			Name:             "Красная Поляна: Горки Город",
			Slug:             "krasnaya-polyana-gorki-gorod",
			DescriptionShort: "Горнолыжный курорт и развлечения круглый год",
			DescriptionFull:  "Горки Город -- курортный комплекс на высоте 960-2200 м. Зимой -- горнолыжные трассы, летом -- треккинг, каньонинг и zip-line. Рестораны, спа и панорамные террасы с видом на Кавказский хребет.",
			Category:         "resort",
			Tags:             []string{"горы", "лыжи", "спорт", "спа", "рестораны"},
			PricePerNight:    15000,
			Capacity:         500,
			AccessLevel:      "open",
			DensityLevel:     "red",
			ChildFriendly:    true,
			Address:          "Краснодарский край, пос. Красная Поляна",
			IsPublished:      true,
			Latitude:         43.6614,
			Longitude:        40.2699,
		},
		{
			Name:             "Кипарисовое озеро",
			Slug:             "kiparisovoe-ozero",
			DescriptionShort: "Болотные кипарисы, растущие прямо из воды",
			DescriptionFull:  "Кипарисовое озеро близ Сукко -- уникальное место, где 32 болотных кипариса растут прямо из воды. Осенью кроны краснеют и создают фантастический пейзаж. Прогулка на лодке, фотографии на рассвете.",
			Category:         "nature",
			Tags:             []string{"озеро", "кипарисы", "фото", "природа", "лодки"},
			PricePerNight:    0,
			Capacity:         100,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    true,
			Address:          "Краснодарский край, с. Сукко",
			IsPublished:      true,
			Latitude:         44.8072,
			Longitude:        37.3889,
		},
		{
			Name:             "Тайная тропа к дольменам",
			Slug:             "taynaya-tropa-k-dolmenam",
			DescriptionShort: "Древние мегалитические сооружения в лесу",
			DescriptionFull:  "Скрытая тропа к группе из 5 дольменов бронзового века (III-II тысячелетие до н.э.) в лесу. Каменные конструкции весом до 30 тонн, покрытые мхом и лианами. Место силы, обнаруженное местными жителями.",
			Category:         "cultural",
			Tags:             []string{"дольмены", "археология", "лес", "тропа", "тайное"},
			PricePerNight:    0,
			Capacity:         5,
			AccessLevel:      "hidden",
			DensityLevel:     "green",
			ChildFriendly:    false,
			Address:          "Краснодарский край, Туапсинский район",
			IsPublished:      true,
			Latitude:         44.0875,
			Longitude:        39.0723,
		},
		{
			Name:             "Рафтинг на реке Белая",
			Slug:             "rafting-na-reke-belaya",
			DescriptionShort: "Сплав по горной реке с порогами 2-3 категории",
			DescriptionFull:  "Рафтинг по реке Белая -- один из лучших маршрутов для водного туризма на юге России. Пороги 2 и 3 категории, живописные каньоны и прозрачная горная вода. Маршрут подходит для новичков и опытных рафтеров.",
			Category:         "extreme",
			Tags:             []string{"рафтинг", "река", "адреналин", "спорт", "каньон"},
			PricePerNight:    0,
			Capacity:         20,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    false,
			Address:          "Республика Адыгея, р. Белая",
			IsPublished:      true,
			Latitude:         44.3000,
			Longitude:        40.1700,
		},
		{
			Name:             "Этнографический комплекс Атамань",
			Slug:             "etnograficheskiy-kompleks-ataman",
			DescriptionShort: "Казачья станица-музей под открытым небом",
			DescriptionFull:  "Атамань -- крупнейший этнографический комплекс юга России. Воссозданная казачья станица с хатами, кузницей, мельницей и подворьями. Народные праздники, кубанский борщ и казачьи песни.",
			Category:         "cultural",
			Tags:             []string{"казаки", "музей", "история", "праздники", "кухня"},
			PricePerNight:    0,
			Capacity:         300,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    true,
			Address:          "Краснодарский край, ст. Тамань",
			IsPublished:      true,
			Latitude:         45.2106,
			Longitude:        36.7114,
		},
		{
			Name:             "Тихая бухта у Прасковеевки",
			Slug:             "tikhaya-bukhta-praskoveyevka",
			DescriptionShort: "Уединённый пляж у скалы Парус",
			DescriptionFull:  "Дикий пляж между Геленджиком и Прасковеевкой. Чистое море, сосновый воздух. Подступает только по тропе через лес (2 км). Нет инфраструктуры -- только природа и тишина. Идеальное место для кемпинга.",
			Category:         "beach",
			Tags:             []string{"пляж", "море", "дикий", "тишина", "кемпинг"},
			PricePerNight:    0,
			Capacity:         15,
			AccessLevel:      "hidden",
			DensityLevel:     "green",
			ChildFriendly:    false,
			Address:          "Краснодарский край, с. Прасковеевка",
			IsPublished:      true,
			Latitude:         44.4300,
			Longitude:        38.1700,
		},
		{
			Name:             "Ферма Кубанский двор",
			Slug:             "ferma-kubanskiy-dvor",
			DescriptionShort: "Агротуризм и кубанская кухня у подножия гор",
			DescriptionFull:  "Аграрная усадьба с собственным хозяйством: коровы, куры, огород. Гости участвуют в сборе яиц, доении коров и приготовлении домашнего кубанского обеда. Два гостевых домика с видом на горы.",
			Category:         "farm",
			Tags:             []string{"ферма", "агротуризм", "кухня", "ночлег", "дети"},
			PricePerNight:    4000,
			Capacity:         6,
			AccessLevel:      "open",
			DensityLevel:     "green",
			ChildFriendly:    true,
			Address:          "Краснодарский край, ст. Нижегородская",
			IsPublished:      true,
			Latitude:         44.5100,
			Longitude:        39.6800,
		},
		{
			Name:             "Орлиная полка",
			Slug:             "orlinaya-polka",
			DescriptionShort: "Скальный выступ с панорамой на три ущелья",
			DescriptionFull:  "Орлиная полка -- скальный козырёк на высоте 1000 м над уровнем моря. Панорамный обзор на три ущелья и горную долину. Маршрут средней сложности (5 км в одну сторону). Фотогеничная точка, ставшая визитной карточкой региона.",
			Category:         "trail",
			Tags:             []string{"горы", "панорама", "треккинг", "фото", "скалы"},
			PricePerNight:    0,
			Capacity:         20,
			AccessLevel:      "open",
			DensityLevel:     "yellow",
			ChildFriendly:    false,
			Address:          "Краснодарский край, пос. Мезмай",
			IsPublished:      true,
			Latitude:         44.2100,
			Longitude:        39.9400,
		},
	}
}
