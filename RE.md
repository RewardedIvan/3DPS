# Yoooo
whas upp??? 
![dog](https://imgs.search.brave.com/h6cXzBdjQLJeVel33pPYteS4IduQxcniXfQM1G_1dME/rs:fit:1200:720:1/g:ce/aHR0cHM6Ly9pLnl0/aW1nLmNvbS92aS95/UXBKdnBjSlR1by9t/YXhyZXNkZWZhdWx0/LmpwZw)
###### The ceiling
### Jokes out the way.....
  

### Every request is sent to http://delugedrop.com/3Dash/whatever with http v1 and your favorite *certified by robtop* php security

# findings lmao
## POST A Level
\> **POST** /push_level_data.php  
\> Content-Type: application/x-www-form-urlencoded  
\># URL Encoded Forms data:  
\> name = "self explanatory"  
\> author = "why not make an account system"
\> difficulty = "0-5"  
\> data = `{"name":"normally the limit is 24 chars","author":"same here","difficulty":0,"songId":0-21,"songStartTime":0-songlength,"floorId":0-3,"backgroundId":0-2,"startingColor":[red,green,blue],"levelData":[#Level data],"pathData":[#Path Data],"cameraData":[#Camera Data]}` 

## GET Recent Levels
\> **GET** /get_recent.php  
< Level ID  
< Level Name  
< Level Difficulty  
< Repeat.............  

## GET A Level
\> **POST** /get_json.php *certified robtop **POST** request*  
\> Content-Type: application/x-www-form-urlencoded  
\># URL Encoded Forms data:  
\> id = 0-yes  
< #Level Data (in json)  

# Datas

## Level Data
### Just look at the name of the function
```csharp
public static int[,] FlatDataToEditorData(List<GameObject>[][] inData, int totalItems)
{
	int[,] array = new int[totalItems, 5];
	int num = 0;
	for (int i = 0; i < inData[0].Length; i++)
	{
		for (int j = 0; j < inData.Length; j++)
		{
			List<GameObject> list = inData[j][i];
			for (int k = 0; k < list.Count; k++)
			{
				FlatItem component = list[k].GetComponent<FlatItem>();
				array[num, 0] = component.index;
				array[num, 1] = component.x;
				array[num, 2] = component.y;
				array[num, 3] = component.z;
				array[num, 4] = component.angle;
				num++;
			}
		}
	}
	return array;
}
```
Yep, one of the many numbers in a single array would look like `[index, x, y, z, angle]`
remove the brackets and continue it with every object and you get the level data

## Path Data
TODO

## Camera Data
```csharp
private void RecordArm()
{
	float[] array = new float[4];
	Vector3 myEulerAngles = this.boomArm.myEulerAngles;
	array[0] = myEulerAngles.x;
	array[1] = myEulerAngles.y;
	array[2] = myEulerAngles.z;
	array[3] = this.time;
	CameraAnimator.recordedPoints.Add(array);
}
```
This function is called every frame update if the "playtester" has not ended and you haven't pressed escape.
This is pretty similar to the #Level Data, but its more like `[x, y, z, time]` and
continue with every angle squish it in to one array and ya done

## Data When Uploading
```csharp
public static Level ExportToLevelObject()
{
	return new Level
	{
		name = LevelEditor.levelName,
		author = LevelEditor.levelAuthor,
		difficulty = LevelEditor.difficulty,
		songId = LevelEditor.songId,
		songStartTime = LevelEditor.songStartTime,
		floorId = LevelEditor.floorId,
		backgroundId = LevelEditor.backgroundId,
		startingColor = LevelEditor.ColorToArray(LevelEditor.startingColor),
		levelData = LevelEditor.GridToArray(LevelEditor.levelData),
		pathData = LevelEditor.GridToArray(LevelEditor.pathData),
		cameraData = LevelEditor.GridToArray(LevelEditor.cameraData)
	};
}
```
This gets converted to JSON
```csharp
public string LevelToJSON(Level level)
{
	return JsonUtility.ToJson(level);
}
```
And gets hand crafted in to a request, also is it just me or did delugedrop forget that the name, author and difficulty is also in the data??
```csharp
private IEnumerator SetRequest(string uri, string levelName, string levelAuthor, int difficulty, string JSON)
{
	WWWForm wwwform = new WWWForm();
	wwwform.AddField("name", levelName);
	wwwform.AddField("author", levelAuthor);
	wwwform.AddField("difficulty", difficulty);
	wwwform.AddField("data", JSON);
	using (UnityWebRequest www = UnityWebRequest.Post(uri, wwwform))
	{
		yield return www.SendWebRequest();
		int responseCode;
		if (www.result == UnityWebRequest.Result.ConnectionError || www.result ==UnityWebRequest.Result.DataProcessingError || www.result ==UnityWebRequest.Result.ProtocolError)
		{
			Debug.Log(www.error);
			responseCode = 0;
		}
		else
		{
			Debug.Log("No Unity Errors");
			responseCode = (int)www.responseCode;
		}
		this.ManageOutput(www.downloadHandler.text, responseCode);
	}
	UnityWebRequest www = null;
	yield break;
	yield break;
}
```

# Security
## Vuln i probably foudn
ok so can i just create mitm proxy, and do something evil 
![big brain](https://en.forum.tribalwars2.com/data/avatars/o/5/5424.jpg?1598231572)
## Havent tested, but
since no server protec i can attac and make ***IMPOSSIBLE*** lvl
![sanitizing user input for the better](https://i.imgur.com/cmp3z6z.png)

## New Absolute for 3Dash????
Ummm, i think yeah unless he kills my work