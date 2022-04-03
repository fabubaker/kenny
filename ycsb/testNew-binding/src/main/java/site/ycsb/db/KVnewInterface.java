package site.ycsb.db;

import site.ycsb.ByteArrayByteIterator;
import site.ycsb.ByteIterator;
import site.ycsb.DBException;

import java.io.IOException;
import java.util.*;


public class KVnewInterface {



    public static void main(String[] args) throws IOException, InterruptedException, DBException {

        CS654KeyValueS client = new CS654KeyValueS();
        client.init();

        //Map<String, ByteIterator> m;


        HashMap<String, ByteIterator> capitalCities = new HashMap<String, ByteIterator>();

        // Add keys and values (Country, City)
        capitalCities.put("England", new ByteArrayByteIterator("London".getBytes()));
        capitalCities.put("Germany", new ByteArrayByteIterator("Berlin".getBytes()));
        capitalCities.put("Norway", new ByteArrayByteIterator("Oslo".getBytes()));
        capitalCities.put("USA", new ByteArrayByteIterator("Washington DC".getBytes()));

        client.insert("table", "city", capitalCities);

        HashMap<String, ByteIterator> updateCities = new HashMap<String, ByteIterator>();
        updateCities.put("Germany", new ByteArrayByteIterator("Russia".getBytes()));
        client.update("table","city", updateCities);



        Set<String> fields = new HashSet<>();
        fields.add("England");
        fields.add("Germany");
        fields.add("Norway");
        fields.add("USA");
        HashMap<String, ByteIterator> newCities = new HashMap<String, ByteIterator>();
        client.read("table", "city", fields, newCities);
        for (String name: newCities.keySet()) {

            String key = name.toString();
            String value = newCities.get(name).toString();
            System.out.println(key + " " + value);
        }


        /*
        HttpClient client = HttpClient.newHttpClient();
        ObjectMapper mapper = new ObjectMapper();
        HashMap<String, String> map = new HashMap<>();
        ArrayList<String> list = new ArrayList<>();
        map.put("field1", "soda");
        map.put("field2", "utr");

        list.add("field2");
        //list.add("field1");

        String json = mapper.writeValueAsString(map);
        String json1 = mapper.writeValueAsString(list);

        String key = "urlKey";

        System.out.println(json);
        HttpRequest request = HttpRequest.newBuilder()
                .PUT(HttpRequest.BodyPublishers.ofString(json))
                .uri(URI.create(String.format("http://localhost:8080/%s", key)))
                .build();

        HttpResponse<String> response =  client.send(request, HttpResponse.BodyHandlers.ofString());



        HttpRequest request1 = HttpRequest.newBuilder()
                .POST(HttpRequest.BodyPublishers.ofString(json1))
                .uri(URI.create(String.format("http://localhost:8080/%s", key)))
                .build();

        HttpResponse<String> response1 =  client.send(request1, HttpResponse.BodyHandlers.ofString());
        System.out.println(mapper.readValue(response1.body(), Map.class));

        Map<String, String> m = mapper.readValue(response1.body(), Map.class);
        System.out.println(m.get("field1"));

        System.out.println(response.body());
        System.out.println(map);
        System.out.println(list);
        System.out.println(mapper.writeValueAsString(map));
        System.out.println(mapper.writeValueAsString(list));
        System.out.println(mapper.readValue(mapper.writeValueAsString(map), Map.class));
        System.out.println(mapper.readValue(response.body(), Map.class));*/

    }

}
